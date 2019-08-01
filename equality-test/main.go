package main

import "fmt"

// JobData contains stripped-down information about a failed job.
type JobData struct {
	jobName     string
	jobType     string
	jobRepo     string
	jobPull     int
	failedTests []string
	flakyTests  []string
}

func main() {
  j1 := JobData{
    jobName: "foo",
    jobType: "bar",
    jobRepo: "baz",
    jobPull: 123,
    failedTests: []string{"test_foo", "test_bar"},
    flakyTests: []string{"test_bar", "test_baz"},
  }

  j2 := JobData{
    jobName: "foo",
    jobType: "bar",
    jobRepo: "baz",
    jobPull: 123,
    failedTests: []string{"test_bar", "test_baz"},
    flakyTests: []string{"test_foo", "test_bar"},
  }

  fmt.Printf("%t\n", j1 == j2)
}
