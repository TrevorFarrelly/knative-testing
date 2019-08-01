package main

import (
  "context"
  "log"
  "encoding/json"
	"cloud.google.com/go/pubsub"
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
    Refs: []Refs{
      {
        Org: "knative",
        Repo: "serving",
        Pulls: []Pull{{
          Number: 4898,
        }},
      },
    },
    JobType: PresubmitJob,
    JobName: "pull-knative-serving-integration-tests",
  }
  msg = testFail
)

func main() {
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
