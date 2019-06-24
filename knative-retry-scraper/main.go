package main

import (
	"log"
  "encoding/json"
  "time"
  "net/http"
  "io/ioutil"
  "strings"
  "fmt"
	"github.com/TrevorFarrelly/test-infra/shared/ghutil"
  "github.com/google/go-github/github"
)

const (
  githubAccount = "/usr/local/google/home/trevorfarrelly/.ssh/githubtoken"
	org = "knative"
	repo = "serving"
  format = "2006-01-02"
  numSamples = 1 // number of samples to collect, in weeks
)
var (
  cutoffTime = numSamples * 7 * 24
  timeInterval = 7 * 24 // 1 week intervals
)

type ParsedComment struct {
  User ParsedUser `json:"user"`
  Body string `json:"body"`
}

type ParsedUser struct {
  User string `json:"login"`
}

type Comment struct {
  User string
  Body string
}

type RetestData struct {
  StartDate time.Time
  EndDate time.Time
  NumPRs int
  NumRetests int
}

func auth() *ghutil.GithubClient {
	gc, err := ghutil.NewGithubClient(githubAccount)
	if err != nil {
		log.Fatalf("GitHub auth error: %v\n", err)
	}
	return gc
}

func getCommentBodies(url string) []Comment {
  resp, err := http.Get(url)
  if err != nil {
    log.Printf("failed to fetch comments: %v", err)
    return nil
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Printf("failed to fetch comments: %v", err)
    return nil
  }
  var parsedComments []ParsedComment
  if err = json.Unmarshal(body, &parsedComments); err != nil {
    log.Printf("failed to fetch comments: %v\n%s\n", err, url)
    return nil
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

func getWeek(date time.Time) (time.Time, time.Time) {
  day := int(date.Weekday())
  offset, _ := time.ParseDuration(fmt.Sprintf("-%dh", day * 24))
  sunday := date.Add(offset)
  offset, _ = time.ParseDuration(fmt.Sprintf("-%dh", 6 * 24))
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
  return prs
}

func getPullRequests(gc *ghutil.GithubClient) []*github.PullRequest {
  data, err := ioutil.ReadFile("cache.json")
  if err != nil {
    log.Printf("Could not read cached PR data: %v", err)
    return getPullRequestsFromAPI(gc)
  }
  var prs []*github.PullRequest
  if err = json.Unmarshal(data, prs); err != nil {
    log.Printf("Could not read cached PR data: %v", err)
    return getPullRequestsFromAPI(gc)
  }
  return prs
}

func main() {
  now := time.Now()
  var samples [numSamples]*RetestData

	for _, pr := range getPullRequests(auth()) {
    age := now.Sub(*pr.CreatedAt).Hours()
    if age < float64(cutoffTime) {
      index := int(age / float64(timeInterval))
      if samples[index] == nil {
        start, end := getWeek(*pr.CreatedAt)
        samples[index] = &RetestData{
          StartDate: start,
          EndDate: end,
        }
      }
      samples[index].NumPRs++
      samples[index].NumRetests += len(getCommentBodies(pr.GetCommentsURL()))
    }
	}

  for week, sample := range samples {
    fmt.Printf("Week: %d (%s - %s): %d PRs, %d Retests\n", week, sample.StartDate.Format(format),
      sample.EndDate.Format(format), sample.NumPRs, sample.NumRetests)
  }
}
