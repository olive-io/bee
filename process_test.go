// Copyright 2024 The bee Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bee_test

import (
	"context"
	"testing"

	"github.com/olive-io/bee"
	"github.com/olive-io/bee/process"
	"github.com/olive-io/bpmn/tracing"
)

func TestRuntime_Play(t *testing.T) {
	sources := []string{"host1"}
	rt, inventory, cancel := newRuntime(t)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	inventory.AddSources(sources...)

	pr := &process.Process{
		Name:  "a test process",
		Id:    "p1",
		Hosts: sources,
		Tasks: []process.ITask{
			&process.Task{
				Name:   "first task",
				Id:     "t1",
				Action: "ping",
			},
		},
	}
	err := rt.Play(ctx, pr, options...)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRuntime_PlayWithTracer(t *testing.T) {
	sources := []string{"host1"}
	rt, inventory, cancel := newRuntime(t)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	tracer := make(chan tracing.ITrace, 10)
	options = append(options, bee.WithRunTracer(tracer))
	inventory.AddSources(sources...)

	pr := &process.Process{
		Name:  "a test process",
		Id:    "p1",
		Hosts: sources,
		Tasks: []process.ITask{
			&process.Task{
				Name:   "first task",
				Id:     "t1",
				Action: "ping",
			},
		},
	}
	
	go func() {
		for tt := range tracer {
			t.Logf("%#v", tt)
		}
	}()
	err := rt.Play(ctx, pr, options...)
	if err != nil {
		t.Fatal(err)
	}
}
