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

type Process struct {
	Name string `json:"name" yaml:"name"`
	Id   string `json:"id" yaml:"id"`

	Hosts []string `json:"hosts" yaml:"hosts"`

	Vars map[string]any `json:"vars" yaml:"vars"`

	RemoteUser string `json:"remote_user" yaml:"remote_user"`

	Sudo     bool   `json:"sudo" yaml:"sudo"`
	SudoUser string `json:"sudo_user" yaml:"sudo_user"`

	Tasks []*Task `json:"tasks" yaml:"tasks" yaml:"tasks"`

	Handlers []*Handler `json:"handlers" yaml:"handlers"`
}

func (p *Process) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var kvs []YamlKV
	if err := unmarshal(&kvs); err != nil {
		return err
	}
	for _, kv := range kvs {
		_, err = kv.Apply("name", &p.Name)
		if err != nil {
			return
		}
		_, err = kv.ApplyArray("hosts", &p.Hosts)
		if err != nil {
			return
		}
		_, err = kv.ApplyMap("vars", &p.Vars)
		if err != nil {
			return
		}
		_, err = kv.Apply("remote_user", &p.RemoteUser)
		if err != nil {
			return
		}
		_, err = kv.Apply("sudo", &p.Sudo)
		if err != nil {
			return
		}
		_, err = kv.Apply("sudo_user", &p.SudoUser)
		if err != nil {
			return
		}
		if values, ok := kv["tasks"]; ok {
			vv, ok := values.([]any)
			if !ok {
				continue
			}
			p.Tasks = make([]*Task, len(vv))
			for i, item := range vv {
				task := new(Task)
				if err = task.fromKV(item.(YamlKV)); err != nil {
					return
				}
				p.Tasks[i] = task
			}
		}
		if values, ok := kv["handlers"]; ok {
			vv, ok := values.([]any)
			if !ok {
				continue
			}
			p.Handlers = make([]*Handler, len(vv))
			for i, item := range vv {
				handler := new(Handler)
				if err = handler.fromKV(item.(YamlKV)); err != nil {
					return
				}
				p.Handlers[i] = handler
			}
		}
	}

	return nil
}
