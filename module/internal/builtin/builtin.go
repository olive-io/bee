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

package builtin

import (
	"github.com/olive-io/bee/module"
	"github.com/olive-io/bee/module/internal"
	"github.com/olive-io/bee/module/internal/builtin/ping"
)

func GetModules() (map[string]*module.Module, error) {
	modules := map[string]*module.Module{}
	if err := register(&modules, ping.PingModule); err != nil {
		return nil, err
	}

	return modules, nil
}

func register(modules *map[string]*module.Module, m *module.Module) error {
	if err := injectScript(m.Command); err != nil {
		return err
	}
	for i := range m.Commands {
		if err := injectScript(m.Commands[i]); err != nil {
			return nil
		}
	}
	(*modules)[m.Name] = m
	return nil
}

func injectScript(c *module.Command) error {
	if !c.Runnable() {
		return nil
	}
	data, err := internal.ReadFile(c.Script)
	if err != nil {
		return err
	}
	if c.ScriptsData == nil {
		c.ScriptsData = map[string][]byte{}
	}
	c.ScriptsData[c.Script] = data
	return nil
}
