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
	"fmt"
)

type Handler struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Id   string `json:"id,omitempty" yaml:"id,omitempty"`
	Desc string `json:"desc,omitempty" yaml:"desc,omitempty"`
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Action string         `json:"action,omitempty" yaml:"action,omitempty"`
	Args   map[string]any `json:"args,omitempty" yaml:"args,omitempty"`
}

func (h *Handler) fromKV(kv YamlKV) (err error) {
	for key, value := range kv {
		if key == "name" {
			_, err = kv.Apply("name", &h.Name)
			if err != nil {
				return err
			}
			continue
		}
		if key == "id" {
			_, err = kv.Apply("id", &h.Id)
			if err != nil {
				return err
			}
			continue
		}
		if key == "desc" {
			_, err = kv.Apply("desc", &h.Desc)
			if err != nil {
				return err
			}
			continue
		}
		if key == "action" {
			_, err = kv.Apply("action", &h.Action)
			if err != nil {
				return
			}
			continue
		}
		if key == "kind" {
			_, err = kv.Apply("kind", &h.Kind)
			if err != nil {
				return
			}
			continue
		}

		if key == "" {
			h.Action = key
		}
		if vs, ok := value.(string); ok {
			h.Args = parseTaskArgs(vs)
		}
		if ykv, ok := value.(YamlKV); ok {
			h.Args = ykv
		}
	}
	return nil
}

func (h *Handler) String() string {
	return fmt.Sprintf(`{"%s": %s}`, h.Name, h.Action)
}
