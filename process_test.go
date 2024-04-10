/*
   Copyright 2024 The bee Authors

   This library is free software; you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public
   License as published by the Free Software Foundation; either
   version 2.1 of the License, or (at your option) any later version.

   This library is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License along with this library;
*/

package bee_test

import (
	"context"
	"testing"

	"github.com/olive-io/bpmn/tracing"

	"github.com/olive-io/bee"
	"github.com/olive-io/bee/process"
)

func TestRuntime_Play(t *testing.T) {
	sources := []string{"host1"}
	rt, inventory, cancel := newRuntime(t)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	_ = inventory.AddSources(sources...)

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

func TestRuntime_PlayWithService(t *testing.T) {
	sources := []string{"host1"}
	rt, inventory, cancel := newRuntime(t)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	options = append(options)
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
				Args: map[string]any{
					"data": "xxx",
				},
				Catch: &process.Handler{
					Name:   "restart service",
					Kind:   "service",
					Action: "service",
					Args: map[string]any{
						"name":  "httpd",
						"state": "latest",
					},
				},
				Finish: &process.Handler{
					Name:   "restart service",
					Kind:   "service",
					Action: "service",
					Args: map[string]any{
						"name":  "nginx",
						"state": "latest",
					},
				},
			},
			&process.Service{
				Name:   "second task",
				Kind:   "service",
				Id:     "t2",
				Action: "test",
				Args: map[string]any{
					"name":      "lack",
					"text":      "This is an easy text",
					"languages": []string{"Go", "Javascript"},
				},
				Catch: &process.Handler{
					Name:   "restart service",
					Kind:   "service",
					Action: "service",
					Args: map[string]any{
						"name":  "httpd",
						"state": "latest",
					},
				},
				Finish: &process.Handler{
					Name:   "restart service",
					Kind:   "service",
					Action: "service",
					Args: map[string]any{
						"name":  "nginx",
						"state": "latest",
					},
				},
			},
		},
	}
	err := rt.Play(ctx, pr, options...)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRuntime_PlayWithTracer(t *testing.T) {
	sources := []string{"host1", "localhost"}
	rt, inventory, cancel := newRuntime(t)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	tracer := make(chan tracing.ITrace, 10)
	options = append(options, bee.WithRunTracer(tracer), bee.WithRunSync(false))
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
				Hosts:  []string{"localhost"},
				Args:   map[string]any{"data": "timeout"},
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
