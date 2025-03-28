// Copyright 2014 armyant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http_task

import (
	"io"
	"os"

	"github.com/shangzongyu/armyant/task"
	"github.com/shangzongyu/armyant/work"
)

// NewManager Run makes all the requests, prints the summary. It blocks until
// all work is done.
func NewManager(t task.Task) task.WorkManager {
	// append hey's user agent
	this := new(Manager)
	//this.results = make(chan *work.Result, t.N)
	return this
}

type Manager struct {
	results chan *work.Result
	// Writer is where results will be written. If nil, results are written to stdout.
	Writer io.Writer
}

func (m *Manager) writer() io.Writer {
	if m.Writer == nil {
		return os.Stdout
	}
	return m.Writer
}

func (m *Manager) Finish(task task.Task) {
	close(m.results)
	//total := time.Now().Sub(task.Start)
	//work.NewReport(m.writer(), task.N, m.results, "", total).Finalize()
}

func (m *Manager) CreateWork() task.Work {
	w := new(Work)
	w.H2 = false
	w.Timeout = 20

	w.DisableCompression = false
	w.DisableKeepAlives = false
	w.DisableRedirects = false
	w.manager = m
	return w
}
