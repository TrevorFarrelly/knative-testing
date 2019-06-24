package main

import (
  "log"
  "flag"

  prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
  "k8s.io/test-infra/prow/config"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/test-infra/prow/pubsub/reporter"
)

const (
  testPubSubProjectName = "trevorfarrelly-knative-2019"
  testPubSubTopicName = "testing"
  testPubSubRunID = "bla"
)

var (
  opts = map[string]*prowapi.ProwJob {
    "success": &prowapi.ProwJob{
      ObjectMeta: metav1.ObjectMeta{
      	Name: "success",
      	Labels: map[string]string{
      		"prow.k8s.io/pubsub.project": testPubSubProjectName,
          "prow.k8s.io/pubsub.topic":   testPubSubTopicName,
      		"prow.k8s.io/pubsub.runID":   testPubSubRunID,
      	},
        Annotations: map[string]string{},
      },
      Status: prowapi.ProwJobStatus{
      	State: prowapi.SuccessState,
      	URL:   "bla",
      },
      Spec: prowapi.ProwJobSpec{
      	Type: prowapi.PresubmitJob,
      	Job:  "bla",
      	Refs: &prowapi.Refs{
          Org: "trevorfarrelly",
          Repo: "knative-testing",
      		Pulls: []prowapi.Pull{{Number: 20}},
      	},
      },
    },
    "failure": &prowapi.ProwJob{
      ObjectMeta: metav1.ObjectMeta{
      	Name: "failure",
      	Labels: map[string]string{
          "prow.k8s.io/pubsub.project": testPubSubProjectName,
          "prow.k8s.io/pubsub.topic":   testPubSubTopicName,
      		"prow.k8s.io/pubsub.runID":   testPubSubRunID,
      	},
        Annotations: map[string]string{},
      },
      Status: prowapi.ProwJobStatus{
      	State: prowapi.FailureState,
      	URL:   "bla",
      },
      Spec: prowapi.ProwJobSpec{
      	Type: prowapi.PresubmitJob,
      	Job:  "bla",
      	Refs: &prowapi.Refs{
          Org: "trevorfarrelly",
          Repo: "knative-testing",
      		Pulls: []prowapi.Pull{{Number: 20}},
      	},
      },
    },
  }
)

func main() {
  var status = flag.String("status", "", "which status to send to pubsub")
  flag.Parse()

  c := reporter.NewReporter(func() *config.Config {
    return &config.Config{
			ProwConfig: config.ProwConfig{
				Plank: config.Plank{
					JobURLPrefixConfig: map[string]string{"*": "guber/"},
  			},
      },
    }
  })

  pj := opts[*status]
  if pj == nil {
    log.Fatalf("error: ProwJob for status '%s' does not exist.\n", *status)
  }
  log.Printf("TESTING: Pull Number is %d\n", pj.Spec.Refs.Pulls[0].Number)
  if _, err := c.Report(opts[*status]); err != nil {
    log.Printf("Could not publish message: %v", err)
  }


}
