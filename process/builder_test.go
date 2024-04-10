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
