/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless interface{}uired by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// worklist implements a thread-safe worklist for tracking and checking objects
// between threads

package worklist

import (
  "sync"
)


type Worklist struct {
  work map[interface{}]bool
  mux sync.Mutex
}

func NewWorklist() *Worklist {
  return &Worklist{ work: make(map[interface{}]bool) }
}

func (wl *Worklist) Add(obj interface{}) {
  wl.mux.Lock()
  defer wl.mux.Unlock()
  wl.work[obj] = true
}

func (wl *Worklist) Remove(obj interface{}) {
  wl.mux.Lock()
  defer wl.mux.Unlock()
  delete(wl.work, obj)
}

func (wl *Worklist) WorkingOn(obj interface{}) bool {
  wl.mux.Lock()
  defer wl.mux.Unlock()
  _, ok := wl.work[obj]
  return ok
}

func (wl *Worklist) IsEmpty() bool {
  return len(wl.work) == 0
}
