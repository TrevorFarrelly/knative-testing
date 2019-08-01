package main

import (
	"context"
	"github.com/google/go-github/github"
	"log"
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"github.com/knative/test-infra/shared/gcs"
)

const (
	serviceAccount = "gcp.json"
	format        = "2006-01-02"
)

var (
	ctx     = context.Background()
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

func main() {
	if err := gcs.Authenticate(ctx, serviceAccount); err != nil {
		log.Fatalf("auth fail: %v", err)
	}
	files := gcs.ListChildrenFiles(ctx, "trevorfarrelly-knative-2019", "serving-retry-data/")
	csv := ""
	for _, f := range files {
		contents, err := gcs.Read(ctx, "trevorfarrelly-knative-2019", f)
		if err != nil {
			log.Printf("%s - read fail: %v", f, err)
			continue
		}
		var data RetestData
		if err := json.Unmarshal(contents, &data); err != nil {
			log.Printf("unmarshal fail: %v", err)
			continue
		}
		csv += fmt.Sprintf("%s,%d,%d,%d,%s\n", data.Date.Format(format), data.NumPRs, data.NumRetests, data.MaxRetries, *data.MaxRetryPR.HTMLURL)
	}
	ioutil.WriteFile("data.csv", []byte(csv), 0644)
}
