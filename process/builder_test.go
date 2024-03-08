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

package process

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuilder(t *testing.T) {
	p := NewProcessBuilder().
		Named("p1", "test process", "this is long description text").
		SetHosts("task1", "task2").
		SetTasks(
			NewChildProcessBuilder().
				Named("c1", "test child process", "").
				SetTasks(
					NewTaskBuilder().
						Named("t1", "", "").
						SetHosts("task1").
						SetAction("ping", nil).
						Build(),
					NewServiceBuilder().
						Named("t2", "", "").
						SetHosts("task2").
						SetAction("RemoteAction.SetName", map[string]any{"name": "newName"}).
						SetNotify("h1").
						SetCatch(NewHandlerBuilder().
							Named("", "stop service", "").
							SetKind(ServiceKey).
							SetAction("RemoteAction.StopService", map[string]any{"name": "httpd"}).
							Build()).
						Build(),
				).
				Build(),
		).
		SetHandlers(
			NewHandlerBuilder().
				Named("h1", "restart service", "").
				SetKind(ServiceKey).
				SetAction("RemoteAction.RestartService", map[string]any{"name": "httpd"}).
				Build(),
		).
		Build()

	data, _ := yaml.Marshal(p)
	t.Log(string(data))
}
