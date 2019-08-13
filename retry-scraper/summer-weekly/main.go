package main

import (
	"context"
	"fmt"
	"knative.dev/test-infra/shared/ghutil"
	"github.com/google/go-github/github"
	"log"
	"net/http"
	"strings"
	"time"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	"io/ioutil"
)

const (
	githubAccount = "/usr/local/google/home/trevorfarrelly/.ssh/githubtoken"
	serviceAccount = "/usr/local/google/home/trevorfarrelly/.ssh/gcp.json"
	queryTemplate = "repo:%s/%s+is:pr+is:merged+merged:>%sT00:00:00&sort=created&order=asc"
	org           = "knative"
	repo          = "serving"
	format        = "2006-01-02"
)

var (
	ctx     = context.Background()
	StartTime, _ = time.Parse("2006-01-02 15:04:05", "2019-05-12 00:00:00") // 12 May 2019
	interval, _ = time.ParseDuration("604800000ms") // 1 week in ms
	EndTime = StartTime.Add(interval)
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
	NumPRs     int                 `json:"prs"`
	NumRetests int                 `json:"retries"`
	MaxRetries int                 `json:"record_ct"`
	MaxRetryPR *github.PullRequest `json:"record_pr"`
}

type PRsearch struct {
	total int
	incomplete bool
	results []github.PullRequest
}

// authorize our accounts
func auth() (*ghutil.GithubClient, *storage.Client) {
	github, err := ghutil.NewGithubClient(githubAccount)
	if err != nil {
		log.Fatalf("GitHub auth error: %v\n", err)
	}
	cloud, err := storage.NewClient(ctx, option.WithCredentialsFile(serviceAccount))
	if err != nil {
		log.Fatalf("Cloud auth error: %v\n", err)
	}
	return github, cloud
}

// get comments from a comment URL
func getCommentBodies(gc *ghutil.GithubClient, url string) []Comment {
	fmt.Println("Querying Comment API...")
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
	fmt.Println("Done.")
	return filterComments(comments)
}

// filter out bot comments
func filterComments(comments []Comment) []Comment {
	var newComments []Comment
	for _, comment := range comments {
		if comment.User != "knative-prow-robot" && comment.User != "knative-flaky-test-reporter-robot" && (strings.Contains(comment.Body, "/test pull-knative-serving-integration-tests") || strings.Contains(comment.Body, "/retest")) {
			newComments = append(newComments, comment)
		}
	}
	return newComments
}

// get 24 hours of PRs merged in knative/serving
func getPullRequests(gc *ghutil.GithubClient) []*github.PullRequest {
	fmt.Println("Querying PR API...")
	prs, _ := gc.ListPullRequests("knative", "serving", "", "")
	fmt.Println("Done.")
	var filtered []*github.PullRequest
	for _, pr := range prs {
		if pr.MergedAt != nil && pr.MergedAt.After(StartTime) {
			filtered = append(filtered, pr)
		}
	}
	return filtered
}

// store logs in our gcs bucket
func writeToCSV(data map[string]*RetestData) {
	csv := ""
	for k, v := range data {
		csv += fmt.Sprintf("%s,%d,%d,%0.3f", k, v.NumPRs, v.NumRetests, float64(v.NumRetests)/float64(v.NumPRs))
		if v.MaxRetryPR != nil && v.MaxRetryPR.HTMLURL != nil {
			csv += fmt.Sprintf(",%s\n", *v.MaxRetryPR.HTMLURL)
		} else {
			csv += fmt.Sprintf(",<nil>\n")
		}
	}
	ioutil.WriteFile("data.csv", []byte(csv), 0644)
}

func main() {
	github, _ := auth()
	data := map[string]*RetestData{}
	prs := getPullRequests(github)
	for EndTime.Before(time.Now()) {
		day := StartTime.Format(format)
		// get all PRs merged since July 1st
		for _, pr := range prs {
			if pr.MergedAt.After(StartTime) && pr.MergedAt.Before(EndTime) {
				// get all comments with retests
				retries := len(getCommentBodies(github, pr.GetCommentsURL()))
				if _, ok := data[day]; !ok {
					data[day] = &RetestData{}
				}
				data[day].NumPRs++
				data[day].NumRetests += retries
				if retries > data[day].MaxRetries {
					data[day].MaxRetries = retries
					data[day].MaxRetryPR = pr
				}
			}
		}
		StartTime = EndTime
		EndTime = StartTime.Add(interval)
	}
	writeToCSV(data)
}
