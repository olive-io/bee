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

package modules

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

var (
	defaultFlags = []string{"bee.yml", "bee.yaml"}

	ErrEmptyDir = errors.New("empty directory")
)

type Module struct {
	*Command

	Dir string
}

func LoadDir(name string) (*Module, error) {
	command, err := readYML(name)
	if err != nil {
		return nil, err
	}

	m := &Module{
		Command: command,
		Dir:     name,
	}
	return m, nil
}

func readYML(dir string) (*Command, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	bee := ""
	subs := make([]*Command, 0)
	for _, ent := range ents {
		if bee == "" && lo.Contains[string](defaultFlags, ent.Name()) {
			bee = filepath.Join(dir, ent.Name())
			continue
		}
		if ent.IsDir() {
			sub, se := readYML(filepath.Join(dir, ent.Name()))
			if se != nil && !errors.Is(se, ErrEmptyDir) {
				return nil, se
			}
			if sub != nil {
				subs = append(subs, sub)
			}
		}
	}
	if bee == "" {
		return nil, ErrEmptyDir
	}

	data, err := os.ReadFile(bee)
	if err != nil {
		return nil, err
	}

	c := new(Command)
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	if err = validateScript(c, dir); err != nil {
		return nil, err
	}

	for _, sub := range c.Commands {
		if err = validateScript(sub, dir); err != nil {
			return nil, err
		}
	}
	c.Commands = append(c.Commands, subs...)
	return c, nil
}

func validateScript(c *Command, dir string) error {
	if c.Script == "" {
		return nil
	}

	if c.Script[0] != '/' {
		c.Script = filepath.Join(dir, c.Script)
	}
	_, err := os.Stat(c.Script)
	if err != nil {
		return errors.Wrapf(err, "invalid script")
	}
	return nil
}

type Command struct {
	Name     string     `yaml:"name"`
	Long     string     `yaml:"long"`
	Script   string     `yaml:"script"`
	Authors  []string   `yaml:"authors"`
	Example  string     `yaml:"example"`
	Params   []Schema   `yaml:"params"`
	Returns  []Schema   `yaml:"returns"`
	Commands []*Command `yaml:"commands"`
}

func (c *Command) Runnable() bool {
	return c.Script != ""
}

type Schema struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Short       string `yaml:"short"`
	Description string `yaml:"description"`
	Example     string `yaml:"example"`
}
