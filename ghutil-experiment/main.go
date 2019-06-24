package main

import (
	"log"
	"github.com/knative/test-infra/shared/ghutil"
)

const (
  githubAccount = "/usr/local/google/home/trevorfarrelly/.ssh/githubtoken"
	org = "TrevorFarrelly"
	repo = "knative-testing"
)

func auth() *ghutil.GithubClient {
	gc, err := ghutil.NewGithubClient(githubAccount)
	if err != nil {
		log.Fatalf("GitHub auth error: %v\n", err)
	}
	log.Println("GitHub auth success")
	return gc
}

func main() {
	gc := auth()
	prs, err := gc.ListPullRequests(org, repo, "", "")
	if err != nil {
		log.Fatalf("error getting pull requests: %v", err)
	}
	for _, pr := range prs {
		log.Printf("adding comment to PR #%d, \"%s\"", pr.GetNumber(), pr.GetTitle())
		_, err := gc.CreateComment(org, repo, pr.GetNumber(), "test comment generated automatically.")
		if err != nil {
			log.Printf("error posting comment: %v")
		}
	}
}
