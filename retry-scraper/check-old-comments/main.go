package main

import (
	"context"
	"log"
	"flag"
	"strings"

	"knative.dev/test-infra/shared/ghutil"
	"github.com/google/go-github/github"
)
const (
	org = "knative"
	repo = "serving"
	botID = 48565599 // knative test reporter robot user ID
)
var ctx = context.Background()

// get retryer comment, if it exists, from PR
func getRetryerComment(c *ghutil.GithubClient, pr *github.PullRequest) (*github.IssueComment, error){
	comments, err := c.ListComments(org, repo, pr.GetNumber())
	if err != nil {
		return nil, err
	}
	return filterComments(comments), nil
}
// find bot comment in list of comments on PR
func filterComments(comments []*github.IssueComment) *github.IssueComment {
	for _, comment := range comments {
		if comment.GetUser().GetID() == botID && strings.Contains(comment.GetBody(), "The following tests are currently flaky. Running them again to verify...") {
			return comment
		}
	}
	return nil
}
// get state of a pull request, either open, closed, or merged
func getPRState(pr *github.PullRequest) string {
	state := pr.GetState()
	if state == "closed" && !pr.GetMergedAt().IsZero() {
		state = "merged"
	}
	return state
}

func main() {
	githubToken := flag.String("github-account", "", "Github token file")
	flag.Parse()

	client, err := ghutil.NewGithubClient(*githubToken)
	if err != nil {
		log.Fatalf("could not register github acct: %v", err)
	}

	log.Printf("Querying PR API...\n")
	prs, err := client.ListPullRequests(org, repo, "", "")
	if err != nil {
		log.Fatalf("Could not get pull requests: %v", err)
	}
	log.Printf("Done querying PR API\n")

	count := 0
	var open []string

	log.Printf("Querying Comment API\n")
	for _, pr := range prs {
		if pr.GetState() != "open" {
			continue
		}
		comment, err := getRetryerComment(client, pr)
		if err != nil || comment == nil {
			continue
		}
		count++
		open = append(open, pr.GetHTMLURL())
	}
	log.Printf("Done querying Comment API\n")
	log.Printf("Got %d open PRs with old comment format:\n", count)
	for _, url := range open {
		log.Println(url)
	}
}
