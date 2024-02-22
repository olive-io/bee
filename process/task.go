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

const (
	ChildProcessKey = "process"
	ServiceKey      = "service"
	TaskKey         = "task"
)

type ITask interface {
	fromKV(kv YamlKV) (err error)
}

type INamedTask interface {
	GetName() string
	GetId() string
}

type ICatchTask interface {
	GetHosts() []string
	GetCatch() *Handler
	GetFinish() *Handler
}

type ChildProcess struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Id   string `json:"id,omitempty" yaml:"id,omitempty"`
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`

	Vars map[string]any `json:"vars,omitempty" yaml:"vars,omitempty"`

	RemoteUser string `json:"remote_user,omitempty" yaml:"remote_user,omitempty"`

	Sudo     bool   `json:"sudo,omitempty" yaml:"sudo,omitempty"`
	SudoUser string `json:"sudo_user,omitempty" yaml:"sudo_user,omitempty"`

	Tasks []ITask `json:"tasks,omitempty" yaml:"tasks,omitempty"`

	Handlers []*Handler `json:"handlers,omitempty" yaml:"handlers,omitempty"`
}

func (c *ChildProcess) fromKV(kv YamlKV) (err error) {
	c.Kind = ChildProcessKey
	for key, value := range kv {
		if key == "name" {
			_, err = kv.Apply("name", &c.Name)
			if err != nil {
				return err
			}
			continue
		}
		if key == "id" {
			_, err = kv.Apply("id", &c.Id)
			if err != nil {
				return err
			}
			continue
		}
		if key == "kind" {
			continue
		}
		if key == "hosts" {
			_, err = kv.ApplyArray("hosts", &c.Hosts)
			if err != nil {
				return
			}
			continue
		}
		if key == "vars" {
			_, err = kv.ApplyMap("vars", &c.Vars)
			if err != nil {
				return
			}
			continue
		}
		if key == "remote_user" {
			_, err = kv.Apply("remote_user", &c.RemoteUser)
			if err != nil {
				return
			}
			continue
		}
		if key == "sudo" {
			_, err = kv.Apply("sudo", &c.Sudo)
			if err != nil {
				return
			}
			continue
		}
		if key == "sudo_user" {
			_, err = kv.Apply("sudo_user", &c.SudoUser)
			if err != nil {
				return
			}
			continue
		}

		if key == "tasks" {
			vv, ok := value.([]any)
			if !ok {
				continue
			}
			c.Tasks = make([]ITask, len(vv))
			for i, item := range vv {

				var kind string
				ykv := item.(YamlKV)
				if exists, _ := ykv.Apply("kind", &kind); exists {
					switch kind {
					case ChildProcessKey:
						cp := new(ChildProcess)
						if err = cp.fromKV(ykv); err != nil {
							return
						}
						c.Tasks[i] = cp
					case ServiceKey:
						sv := new(Service)
						if err = sv.fromKV(ykv); err != nil {
							return
						}
						c.Tasks[i] = sv
					}

				} else {
					task := new(Task)
					if err = task.fromKV(ykv); err != nil {
						return
					}
					c.Tasks[i] = task
				}
			}
		}

		if values, ok := kv["handlers"]; ok {
			vv, ok := values.([]any)
			if !ok {
				continue
			}
			c.Handlers = make([]*Handler, len(vv))
			for i, item := range vv {
				handler := new(Handler)
				if err = handler.fromKV(item.(YamlKV)); err != nil {
					return
				}
				c.Handlers[i] = handler
			}
		}
	}
	return
}

type Task struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Id   string `json:"id,omitempty" yaml:"id,omitempty"`

	Vars map[string]any `json:"vars,omitempty" yaml:"vars,omitempty"`

	Action string         `json:"action,omitempty" yaml:"action,omitempty"`
	Args   map[string]any `json:"args,omitempty" yaml:"args,omitempty"`

	RemoteUser string `json:"remote_user,omitempty" yaml:"remote_user,omitempty"`

	Sudo     bool   `json:"sudo,omitempty" yaml:"sudo,omitempty"`
	SudoUser string `json:"sudo_user,omitempty" yaml:"sudo_user,omitempty"`

	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`

	Catch  *Handler `json:"catch,omitempty" yaml:"catch,omitempty"`
	Finish *Handler `json:"finish,omitempty" yaml:"finish,omitempty"`

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
			_, err = kv.Apply("id", &t.Id)
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
		if key == "kind" {
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

		if key == "catch" {
			if vv, ok := value.(YamlKV); ok {
				t.Catch = new(Handler)
				if err = t.Catch.fromKV(vv); err != nil {
					return err
				}
			}
			continue
		}

		if key == "finish" {
			if vv, ok := value.(YamlKV); ok {
				t.Finish = new(Handler)
				if err = t.Finish.fromKV(vv); err != nil {
					return err
				}
			}
			continue
		}

		if key == "action" {
			_, err = kv.Apply("action", &t.Action)
			if err != nil {
				return
			}
			continue
		}

		if t.Action == "" {
			t.Action = key
		}
		if vs, ok := value.(string); ok {
			t.Args = parseTaskArgs(vs)
		}
		if ykv, ok := value.(YamlKV); ok {
			t.Args = ykv
		}
	}
	return
}

func (t *Task) GetName() string {
	return t.Name
}

func (t *Task) GetId() string {
	return t.Id
}

func (t *Task) GetHosts() []string {
	return t.Hosts
}

func (t *Task) GetCatch() *Handler {
	return t.Catch
}

func (t *Task) GetFinish() *Handler {
	return t.Finish
}

type Service struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Id   string `json:"id,omitempty" yaml:"id,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Vars map[string]any `json:"vars,omitempty" yaml:"vars,omitempty"`

	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`

	Action string         `json:"action,omitempty" yaml:"action,omitempty"`
	Args   map[string]any `json:"args,omitempty" yaml:"args,omitempty"`

	Catch  *Handler `json:"catch,omitempty" yaml:"catch,omitempty"`
	Finish *Handler `json:"finish,omitempty" yaml:"finish,omitempty"`

	Notify []string `json:"notify,omitempty" yaml:"notify,omitempty"`
}

func (s *Service) fromKV(kv YamlKV) (err error) {
	s.Kind = ServiceKey
	for key, value := range kv {
		if key == "name" {
			_, err = kv.Apply("name", &s.Name)
			if err != nil {
				return err
			}
			continue
		}
		if key == "id" {
			_, err = kv.Apply("id", &s.Name)
			if err != nil {
				return err
			}
			continue
		}
		if key == "hosts" {
			_, err = kv.ApplyArray("hosts", &s.Hosts)
			if err != nil {
				return
			}
			continue
		}
		if key == "vars" {
			_, err = kv.ApplyMap("vars", &s.Vars)
			if err != nil {
				return
			}
			continue
		}
		if key == "notify" {
			_, err = kv.ApplyArray("notify", &s.Notify)
			if err != nil {
				return
			}
			continue
		}
		if key == "kind" {
			continue
		}

		if key == "catch" {
			if vv, ok := value.(YamlKV); ok {
				s.Catch = new(Handler)
				if err = s.Catch.fromKV(vv); err != nil {
					return err
				}
			}
			continue
		}

		if key == "finish" {
			if vv, ok := value.(YamlKV); ok {
				s.Finish = new(Handler)
				if err = s.Finish.fromKV(vv); err != nil {
					return err
				}
			}
			continue
		}

		if key == "action" {
			_, err = kv.Apply("action", &s.Action)
			if err != nil {
				return
			}
			continue
		}

		if key == "" {
			s.Action = key
		}
		if vs, ok := value.(string); ok {
			s.Args = parseTaskArgs(vs)
		}
		if ykv, ok := value.(YamlKV); ok {
			s.Args = ykv
		}
	}
	return
}

func (s *Service) GetName() string {
	return s.Name
}

func (s *Service) GetId() string {
	return s.Id
}

func (s *Service) GetHosts() []string {
	return s.Hosts
}

func (s *Service) GetCatch() *Handler {
	return s.Catch
}

func (s *Service) GetFinish() *Handler {
	return s.Finish
}
