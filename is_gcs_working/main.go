package main

import (
  "log"

  "github.com/TrevorFarrelly/test-infra/shared/prow"
)

func main() {
  prow.Initialize("gcp.json")

  job := prow.NewJob("ci-knative-flakes-reporter", "periodic", "", 0)
  build := job.NewBuild(1149287488397774849)

  for _, artifact := range build.GetArtifacts() {
    log.Printf("%s\n", artifact)
  }
}
