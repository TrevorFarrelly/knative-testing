package main

import (
  "context"
  "log"
  "encoding/json"
	"cloud.google.com/go/pubsub"
  "flag"
)

const (
  testPubSubProjectName = "trevorfarrelly-knative-2019"
  testPubSubTopicName = "test-infra-monitoring-sub"
)

var (
  testFail = &ReportMessage{
    Project: testPubSubProjectName,
    Topic: testPubSubTopicName,
    RunID: "foo",
    Status: FailureState,
    URL: "bar",
    GCSPath: "gs://knative-prow/pr-logs/pull/knative_serving/5106/pull-knative-serving-integration-tests/1159852250034606080",
    Refs: []Refs{
      {
        Org: "knative",
        Repo: "serving",
        Pulls: []Pull{{
          Number: 5006,
        }},
      },
    },
    JobType: PresubmitJob,
    JobName: "pull-knative-serving-integration-tests",
  }
)

func makeMsg(repo, jobName, jobType, status string) *ReportMessage {
  msg := testFail
  msg.JobName = jobName
  msg.JobType = ProwJobType(jobType)
  msg.Status = ProwJobState(status)
  msg.Refs[0].Repo = repo
  return testFail
}

func main() {
  repo := flag.String("repo", "serving", "")
  jobType := flag.String("type", string(PresubmitJob), "")
  status := flag.String("status", string(FailureState), "")
  flag.Parse()
  msg := makeMsg(*repo, "pull-knative-serving-integration-tests", *jobType, *status)
  ctx := context.Background()
  client, err := pubsub.NewClient(ctx, msg.Project)
  if err != nil {
    log.Fatalf("failed to create client")
  }
  topic := client.Topic(msg.Topic)
  d, err := json.Marshal(msg)
  if err != nil {
    log.Fatalf("failed to marshal message")
  }
  res := topic.Publish(ctx, &pubsub.Message{
    Data: d,
  })
  _, err = res.Get(ctx)
  if err != nil {
    log.Fatalf("failed to publish message")
  }
}
