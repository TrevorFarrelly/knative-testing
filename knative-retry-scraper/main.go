package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TrevorFarrelly/test-infra/shared/ghutil"
	"github.com/google/go-github/github"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	githubAccount = "/usr/local/google/home/trevorfarrelly/.ssh/githubtoken"
	org           = "knative"
	repo          = "serving"
	format        = "2006-01-02"
	numSamples    = 12 // number of samples to collect, in weeks
)

var (
	cutoffTime   = numSamples * 7 * 24
	timeInterval = 7 * 24 // 1 week intervals
	ctx = context.Background()
)

type ParsedComment struct {
	User ParsedUser `json:"user"`
	Body string     `json:"body"`
}

type ParsedUser struct {
	User string `json:"login"`
}

type Comment struct {
	User string
	Body string
}

type RetestData struct {
	StartDate  time.Time
	EndDate    time.Time
	NumPRs     int
	NumRetests int
}

func auth() *ghutil.GithubClient {
	gc, err := ghutil.NewGithubClient(githubAccount)
	if err != nil {
		log.Fatalf("GitHub auth error: %v\n", err)
	}
	return gc
}

func getCommentBodies(gc *ghutil.GithubClient, url string) []Comment {
	req, _ := http.NewRequest("GET", url, nil)
	var parsedComments []ParsedComment
	if _, err := gc.Client.Do(ctx, req, &parsedComments); err != nil {
		log.Printf("failed to fetch comments: %v\n%s\n", err, url)
	}
	var comments []Comment
	for _, comment := range parsedComments {
		comments = append(comments, Comment{
			User: comment.User.User,
			Body: comment.Body,
		})
	}
	return filterComments(comments)
}

func filterComments(comments []Comment) []Comment {
	var newComments []Comment
	for _, comment := range comments {
		if comment.User != "knative-prow-robot" && (strings.Contains(comment.Body, "/test") || strings.Contains(comment.Body, "/retest")) {
			newComments = append(newComments, comment)
		}
	}
	return newComments
}

func RateLimited(gc *ghutil.GithubClient) bool {
	req, _ := http.NewRequest("GET", "https://api.github.com/rate_limit", nil)
	var resp struct {
		Rate struct {
			Limit     int   `json:"limit"`
			Remaining int   `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"rate"`
	}
	if _, err := gc.Client.Do(ctx, req, &resp); err != nil {
		log.Printf("rate limit check error: %v", err)
	}
	log.Printf("limit: %d, remaining: %d, reset: %s", resp.Rate.Limit, resp.Rate.Remaining, time.Unix(resp.Rate.Reset, 0).String())
	return resp.Rate.Remaining == 0
}

func getWeek(date time.Time) (time.Time, time.Time) {
	day := int(date.Weekday())
	offset, _ := time.ParseDuration(fmt.Sprintf("-%dh", day*24))
	sunday := date.Add(offset)
	offset, _ = time.ParseDuration(fmt.Sprintf("-%dh", 6*24))
	saturday := sunday.Add(offset)
	return saturday, sunday
}

func getPullRequestsFromAPI(gc *ghutil.GithubClient) []*github.PullRequest {
	prs, err := gc.ListPullRequests(org, repo, "", "")
	if err != nil {
		log.Fatalf("error getting pull requests: %v", err)
	}
	data, err := json.Marshal(prs)
	if err != nil {
		log.Printf("Could not cache PR data: %v", err)
	}
	if err = ioutil.WriteFile("cache.json", data, 0777); err != nil {
		log.Printf("Coult not cache PR data: %v", err)
	}
	log.Printf("Got PR data from API\n")
	return prs
}

func getPullRequests(gc *ghutil.GithubClient) []*github.PullRequest {
	data, err := ioutil.ReadFile("cache.json")
	if err != nil {
		log.Printf("Could not read cached PR data, calling API\n")
		return getPullRequestsFromAPI(gc)
	}
	var prs []*github.PullRequest
	if err = json.Unmarshal(data, &prs); err != nil {
		log.Printf("Could not read cached PR data, calling API\n")
		return getPullRequestsFromAPI(gc)
	}
	log.Printf("Got PR data from cache\n")
	return prs
}

func main() {
	gc := auth()
	if RateLimited(gc) {
		log.Fatalf("Rate-limited, cannot continue")
	}

	now := time.Now()
	var samples [numSamples]*RetestData
	var maxRetries [numSamples]struct{
		count int
		url 	string
	}
	for _, pr := range getPullRequests(gc) {
		age := now.Sub(*pr.CreatedAt).Hours()
		if age < float64(cutoffTime) {
			index := int(age / float64(timeInterval))
			if samples[index] == nil {
				start, end := getWeek(*pr.CreatedAt)
				samples[index] = &RetestData{
					StartDate: start,
					EndDate:   end,
				}
			}
			retries := len(getCommentBodies(gc, pr.GetCommentsURL()))
			samples[index].NumPRs++
			samples[index].NumRetests += retries
			if retries > maxRetries[index].count {
				maxRetries[index].count = retries
			 	maxRetries[index].url = *pr.HTMLURL
			}
		}
	}
	fmt.Printf("Max Retries Each Week:\n")
	for week, data := range maxRetries {
		fmt.Printf("Week %d: %d retries on %s\n", week, data.count, data.url)
	}


	fmt.Printf("Aggregate Week Data:\n")
	output := ""
	for week, sample := range samples {
		fmt.Printf("Week: %d (%s - %s): %d PRs, %d Retests\n", week, sample.StartDate.Format(format),
			sample.EndDate.Format(format), sample.NumPRs, sample.NumRetests)
		output += fmt.Sprintf("%d,%s,%d,%d\n", week, sample.StartDate.Format(format), sample.NumPRs, sample.NumRetests)
	}
	ioutil.WriteFile("data.csv", []byte(output), 0777)
}
