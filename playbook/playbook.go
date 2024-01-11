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

package playbook

import "strings"

type PlayBook struct {
	Name string `json:"name" yaml:"name"`

	Hosts []string `json:"hosts" yaml:"hosts"`

	Vars map[string]any `json:"vars" yaml:"vars"`

	RemoteUser string `json:"remote_user" yaml:"remote_user"`

	Sudo     bool   `json:"sudo" yaml:"sudo"`
	SudoUser string `json:"sudo_user" yaml:"sudo_user"`

	Tasks []*Task `json:"tasks" yaml:"tasks" yaml:"tasks"`

	Handlers []*Handler `json:"handlers" yaml:"handlers"`
}

func (pb *PlayBook) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var kvs []YamlKV
	if err := unmarshal(&kvs); err != nil {
		return err
	}
	for _, kv := range kvs {
		_, err = kv.Apply("name", &pb.Name)
		if err != nil {
			return
		}
		_, err = kv.ApplyArray("hosts", &pb.Hosts)
		if err != nil {
			return
		}
		_, err = kv.ApplyMap("vars", &pb.Vars)
		if err != nil {
			return
		}
		_, err = kv.Apply("remote_user", &pb.RemoteUser)
		if err != nil {
			return
		}
		_, err = kv.Apply("sudo", &pb.Sudo)
		if err != nil {
			return
		}
		_, err = kv.Apply("sudo_user", &pb.SudoUser)
		if err != nil {
			return
		}
		if values, ok := kv["tasks"]; ok {
			vv, ok := values.([]any)
			if !ok {
				continue
			}
			pb.Tasks = make([]*Task, len(vv))
			for i, item := range vv {
				task := new(Task)
				if err = task.fromKV(item.(YamlKV)); err != nil {
					return
				}
				pb.Tasks[i] = task
			}
		}
		if values, ok := kv["handlers"]; ok {
			vv, ok := values.([]any)
			if !ok {
				continue
			}
			pb.Handlers = make([]*Handler, len(vv))
			for i, item := range vv {
				handler := new(Handler)
				if err = handler.fromKV(item.(YamlKV)); err != nil {
					return
				}
				pb.Handlers[i] = handler
			}
		}
	}

	return nil
}

type Task struct {
	Name string `json:"name" yaml:"name"`

	Vars map[string]string `json:"vars" yaml:"vars"`

	Module string            `json:"module" yaml:"module"`
	Args   map[string]string `json:"args" yaml:"args"`

	RemoteUser string `json:"remote_user" yaml:"remote_user"`

	Sudo     bool   `json:"sudo" yaml:"sudo"`
	SudoUser string `json:"sudo_user" yaml:"sudo_user"`

	Hosts []string `json:"hosts" yaml:"hosts"`

	Notify []string `json:"notify" yaml:"notify"`
}

func (t *Task) fromKV(kv YamlKV) (err error) {
	for key, value := range kv {
		if key == "name" {
			_, err = kv.Apply("name", &t.Name)
			if err != nil {
				return err
			}
			continue
		}
		if key == "hosts" {
			_, err = kv.ApplyArray("hosts", &t.Hosts)
			if err != nil {
				return
			}
			continue
		}
		if key == "vars" {
			_, err = kv.ApplyMap("vars", &t.Vars)
			if err != nil {
				return
			}
			continue
		}
		if key == "remote_user" {
			_, err = kv.Apply("remote_user", &t.RemoteUser)
			if err != nil {
				return
			}
			continue
		}
		if key == "sudo" {
			_, err = kv.Apply("sudo", &t.Sudo)
			if err != nil {
				return
			}
			continue
		}
		if key == "sudo_user" {
			_, err = kv.Apply("sudo_user", &t.SudoUser)
			if err != nil {
				return
			}
			continue
		}
		if key == "notify" {
			_, err = kv.ApplyArray("notify", &t.Notify)
			if err != nil {
				return
			}
			continue
		}
		t.Module = key
		if vs, ok := value.(string); ok {
			t.Args = parseTaskArgs(vs)
		}
		if ykv, ok := value.(YamlKV); ok {
			t.Args = map[string]string{}
			for yk, yv := range ykv {
				v0, _ := yv.(string)
				if v0 != "" {
					t.Args[yk] = v0
				}
			}
		}
	}
	return
}

type Handler struct {
	Name string `json:"name" yaml:"name"`

	Module string            `json:"module" yaml:"module"`
	Args   map[string]string `json:"args" yaml:"args"`
}

func (h *Handler) fromKV(kv YamlKV) error {
	for key, value := range kv {
		if key == "name" {
			_, err := kv.Apply("name", &h.Name)
			if err != nil {
				return err
			}
			continue
		}
		h.Module = key
		if vs, ok := value.(string); ok && len(strings.TrimSpace(vs)) != 0 {
			h.Args = parseTaskArgs(vs)
		}
	}
	return nil
}
