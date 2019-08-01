package main

import (
  "time"
  "log"
  "fmt"

  "github.com/TrevorFarrelly/knative-testing/worklist-test/worklist"
)

type testData struct {
  name string
  id int
  bla []float64
}

func (t *testData) String() string {
  return fmt.Sprintf("(%s, %d)", t.name, t.id)
}

func threads(id int, wl *worklist.Worklist) {
  for i := 0; i < 4; i++ {
    log.Printf("TH %d: Adding obj\n", id)
    data := &testData{
      name: fmt.Sprintf("Thread-%d", id),
      id: i,
    }
    wl.Add(data)
    time.Sleep(300 * time.Millisecond)
  }
  //log.Printf("TH %d: contents right now: %s", id, wl.String())
  for i := 0; i < 4; i++ {
    log.Printf("TH %d: Removing obj\n", id)
    data := &testData{
      name: fmt.Sprintf("Thread-%d", id),
      id: i,
    }
    wl.Remove(data)
    time.Sleep(400 * time.Millisecond)
  }

  log.Printf("TH %d: Done", id)
}

func main() {
  wl := worklist.NewWorklist()
  for i := 0; i < 4; i++ {
    log.Printf("MAIN: Spawning thread %d\n", i)
    go threads(i, wl)
    data := &testData{
      name: fmt.Sprintf("main"),
      id: i,
    }
    wl.Add(data)
    time.Sleep(500 * time.Millisecond)
  }
  log.Printf("MAIN: done spawning threads\n")

  log.Printf("MAIN: worklist has finished\n")
}
