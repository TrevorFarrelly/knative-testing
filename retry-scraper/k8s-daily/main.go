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
	"encoding/json"
)

const (
	githubAccount = "/secrets/github"
	serviceAccount = "/secrets/gcp.json"
	queryTemplate = "repo:%s/%s is:pr is:merged merged:>%sT00:00:00"
	org           = "knative"
	repo          = "serving"
	format        = "2006-01-02"
)

var (
	ctx     = context.Background()
	lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
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
	Date       time.Time           `json:"date"`
	NumPRs     int                 `json:"prs"`
	NumRetests int                 `json:"retries"`
	MaxRetries int                 `json:"record_ct"`
	MaxRetryPR *github.PullRequest `json:"record_pr"`
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

// filter out bot comments
func filterComments(comments []Comment) []Comment {
	var newComments []Comment
	for _, comment := range comments {
		if comment.User != "knative-prow-robot" && comment.User != "knative-flaky-test-reporter-robot" && (strings.Contains(comment.Body, "/test") || strings.Contains(comment.Body, "/retest")) {
			newComments = append(newComments, comment)
		}
	}
	return newComments
}

// get 24 hours of PRs merged in knative/serving
func getPullRequests(gc *ghutil.GithubClient) []*github.PullRequest {
	age, _ := time.ParseDuration("-24h")
	query := fmt.Sprintf(queryTemplate, org, repo, time.Now().Add(age).Format(format))
	prs, err := gc.SearchPullRequests(query)
	if err != nil {
		log.Fatalf("error getting pull requests: %v", err)
	}
	return prs
}

// store logs in our gcs bucket
func writeToBucket(cc *storage.Client, data *RetestData) {
	contents, err := json.Marshal(data)
	if err != nil {
		log.Printf("Could not marshal data: %v", err)
	}
	path := "serving-retry-data/" + data.Date.Format(format)
	writer := cc.Bucket("trevorfarrelly-knative-2019").Object(path).NewWriter(ctx)
	if _, err := writer.Write(contents); err != nil {
		log.Printf("Could not write to gcs bucket: %v", err)
	}
	writer.Close()
	log.Printf("wrote report to bucket at %s\n", path)
}

func main() {
	github, cloud := auth()
	for {

		// wait 24 hours for next run
		nextRun := lastRun.AddDate(0, 0, 1)
		time.Sleep(time.Until(nextRun))
		lastRun = nextRun

		today := RetestData{
			Date: time.Now(),
		}

		// get all PRs merged in the last 24 hours
		for _, pr := range getPullRequests(github) {
			// get all comments with retests
			retries := len(getCommentBodies(github, pr.GetCommentsURL()))
			today.NumPRs++
			today.NumRetests += retries
			if retries > today.MaxRetries {
				today.MaxRetries = retries
				today.MaxRetryPR = pr
			}
		}

		writeToBucket(cloud, &today)
		log.Printf("%d PRs, %d Retries. Record Retries: %d on %s\n",
			today.NumPRs, today.NumRetests, today.MaxRetries, *today.MaxRetryPR.HTMLURL)

		writeToBucket(cloud, &today)
	}
}
