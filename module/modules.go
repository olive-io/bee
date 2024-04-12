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

package module

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

var (
	defaultFlags = []string{"bee.yml", "bee.yaml"}

	ErrEmptyDir = errors.New("empty directory")
	ctxValue    = "command"
)

func LoadDir(name string) (*Module, error) {
	command, err := readCommand(name)
	if err != nil {
		return nil, err
	}

	m := &Module{
		Command: *command,
		Dir:     name,
	}

	return m, nil
}

func readCommand(dir string) (*Command, error) {
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
			sub, se := readCommand(filepath.Join(dir, ent.Name()))
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
	if c.Alias == "" {
		c.Alias = c.Name
	}
	c.Run = DefaultRunCommand

	if err = validateScript(c, dir); err != nil {
		return nil, err
	}

	for i := range c.Commands {
		sub := c.Commands[i]
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

	sp := c.Script
	if sp[0] != '/' {
		sp = filepath.Join(dir, c.Script)
	}
	_, err := os.Stat(sp)
	if err != nil {
		return errors.Wrapf(err, "invalid script")
	}
	c.Run = DefaultRunCommand
	return nil
}

type Module struct {
	Command `yaml:",inline"`
	Dir     string `yaml:"dir"`
}

func (m *Module) Execute(args ...string) (*Command, error) {
	cmd := m.Command.ParseCmd()
	if len(args) > 0 &&
		!strings.Contains(args[0], "-") &&
		!strings.Contains(args[0], "=") &&
		strings.Contains(args[0], ".") {
		arg0 := strings.Split(args[0], ".")
		args = append(arg0, args[1:]...)
	}
	for i := range args {
		arg := args[i]
		if strings.Contains(arg, "=") && !strings.HasPrefix(arg, "-") {
			args[i] = "--" + arg
		}
	}
	cmd.SetArgs(args)

	command, err := cmd.ExecuteC()
	if err != nil {
		return nil, err
	}
	mc := command.Context().Value(ctxValue).(*Command)
	mc.Root = m.Root
	return mc, nil
}
