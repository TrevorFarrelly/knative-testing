package main

import (
  "context"
  "log"
  "github.com/knative/test-infra/shared/ghutil"
)

const (
  proj = "trevorfarrelly-knative-2019"
  name = "testing"
)

func auth() *ghutil.GithubClient {
	gc, err := ghutil.NewGithubClient("/usr/local/google/home/trevorfarrelly/.ssh/githubtoken")
	if err != nil {
		log.Fatalf("GitHub auth error: %v\n", err)
	}
	log.Println("GitHub auth success")
	return gc
}

func analyze(msg* ReportMessage) {
  if msg.Status == "failure" {
    log.Printf("Job failed. Posting comment to PR #%d\n", msg.Refs[0].Pulls[0].Number)
    _, err := auth().CreateComment(msg.Refs[0].Org, msg.Refs[0].Repo, msg.Refs[0].Pulls[0].Number, "This comment was created by a failed Prow job")
    if err != nil {
      log.Printf("error posting comment: %v")
    }

  } else {
    log.Printf("Job changed state, but no failure. Continuing...")
  }
}

func main() {

  ctx := context.Background()

  c, err := NewSubscriberClient(ctx, proj, name)
  if err != nil {
    log.Fatalf("could not create client: %v", err)
  }

  log.Printf("listening for pubsub messages...\n")
  for {
    err := c.ReceiveMessageAckAll(ctx, func (msg* ReportMessage) {
      go analyze(msg)
    })
    if err != nil {
      log.Printf("recv failed: %v", err)
    }
  }
}
