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
