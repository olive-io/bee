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

type Task struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Id   string `json:"id,omitempty" yaml:"id,omitempty"`

	Vars map[string]string `json:"vars,omitempty" yaml:"vars,omitempty"`

	Action string            `json:"action,omitempty" yaml:"action,omitempty"`
	Args   map[string]string `json:"args,omitempty" yaml:"args,omitempty"`

	RemoteUser string `json:"remote_user,omitempty" yaml:"remote_user,omitempty"`

	Sudo     bool   `json:"sudo,omitempty" yaml:"sudo,omitempty"`
	SudoUser string `json:"sudo_user,omitempty" yaml:"sudo_user,omitempty"`

	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`

	Notify []string `json:"notify,omitempty" yaml:"notify,omitempty"`
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
		if key == "id" {
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
		t.Action = key
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
