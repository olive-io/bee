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

import "strings"

type Handler struct {
	Name string `json:"name" yaml:"name"`

	Action string         `json:"action,omitempty" yaml:"action,omitempty"`
	Args   map[string]any `json:"args,omitempty" yaml:"args,omitempty"`
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
		h.Action = key
		if vs, ok := value.(string); ok && len(strings.TrimSpace(vs)) != 0 {
			h.Args = parseTaskArgs(vs)
		}
	}
	return nil
}
