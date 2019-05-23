package main

import (
	"log"
)

const (
  githubAccount = "/usr/local/google/home/trevorfarrelly/.ssh/githubtoken"
)

func main() {
  gc, err := NewGithubClient(githubAccount)
  if err != nil {
    log.Fatalf("GitHub auth error: %v\n", err)
  }
  log.Println("GitHub auth success")

  _, err = gc.GetGithubUser()
  if err != nil {
    log.Fatalf("GitHub user error: %v\n", err)
  }
  log.Println("got GitHub user")

  repos, err := gc.ListRepos("TrevorFarrelly")
  if err != nil {
    log.Printf("unable to get repos: %v\n", err)
  } else {
    log.Printf("Found repo data:\n%s\n\n", repos)
  }

  _, err = gc.CreateIssue("TrevorFarrelly", "knative-testing", "test issue", "created from test github client")
  if err != nil {
    log.Printf("Unable to create issue: %v\n", err)
  }
}
