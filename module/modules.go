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

package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	defaultFlags = []string{"bee.yml", "bee.yaml"}

	ErrEmptyDir = errors.New("empty directory")
	ctxValue    = "command"
)

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
	data, err := os.ReadFile(c.Script)
	if err != nil {
		return errors.Wrapf(err, "invalid script")
	}
	if c.ScriptsData == nil {
		c.ScriptsData = map[string][]byte{}
	}
	c.ScriptsData[c.Script] = data
	return nil
}

type Module struct {
	*Command

	Dir string
}

func (m *Module) Execute(patten string) (*Command, error) {
	cmd := m.Command.ParseCmd()
	args := strings.Split(patten, " ")
	if len(args) > 0 &&
		!strings.Contains(args[0], ".") &&
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

	cc, err := cmd.ExecuteC()
	if err != nil {
		return nil, err
	}
	c := cc.Context().Value(ctxValue).(*Command)
	return c, nil
}

type Command struct {
	Name        string            `yaml:"name"`
	Long        string            `yaml:"long"`
	Script      string            `yaml:"script"`
	Authors     []string          `yaml:"authors"`
	Version     string            `yaml:"version"`
	Example     string            `yaml:"example"`
	Params      []*Schema         `yaml:"params"`
	Returns     []*Schema         `yaml:"returns"`
	Commands    []*Command        `yaml:"commands"`
	ScriptsData map[string][]byte `yaml:"-"`
	cobra       *cobra.Command    `yaml:"-"`
}

func (c *Command) Runnable() bool {
	return c.Script != ""
}

func (c *Command) ParseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           c.Name,
		Long:          c.Long,
		Example:       c.Example,
		Annotations:   map[string]string{"script": c.Script},
		Version:       c.Version,
		RunE:          func(cmd *cobra.Command, args []string) error { return nil },
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	flags := cmd.PersistentFlags()
	for _, param := range c.Params {
		pv := param.InitValue()
		_ = pv.Set(param.Default)
		flags.VarP(pv, param.Name, param.Short, param.Description)
	}
	for _, sc := range c.Commands {
		sub := sc.ParseCmd()
		cmd.AddCommand(sub)
	}
	ctx := context.WithValue(context.Background(), ctxValue, c)
	cmd.SetContext(ctx)
	c.cobra = cmd
	return cmd
}

func (c *Command) Flags() *pflag.FlagSet {
	return c.cobra.PersistentFlags()
}

type Schema struct {
	Name        string       `yaml:"name"`
	Type        string       `yaml:"type"`
	Short       string       `yaml:"short"`
	Description string       `yaml:"description"`
	Default     string       `yaml:"default"`
	Example     string       `yaml:"example"`
	Value       *SchemaValue `yaml:"-"`
}

func (s *Schema) InitValue() *SchemaValue {
	sv := &SchemaValue{}
	switch s.Type {
	case "string":
		sv.StringP = new(string)
	case "int", "int32", "int64":
		sv.IntP = new(int64)
	case "uint", "uint32", "uint64":
		sv.UintP = new(uint64)
	case "float", "float32", "float64":
		sv.FloatP = new(float64)
	case "duration":
		sv.DurationP = new(time.Duration)
	}

	s.Value = sv
	return sv
}

type SchemaValue struct {
	IntP      *int64
	UintP     *uint64
	StringP   *string
	FloatP    *float64
	DurationP *time.Duration
}

func (sv *SchemaValue) String() string {
	if sv.IntP != nil {
		return fmt.Sprintf("%d", *sv.IntP)
	}
	if sv.UintP != nil {
		return fmt.Sprintf("%d", &sv.UintP)
	}
	if sv.StringP != nil {
		return *sv.StringP
	}
	if sv.FloatP != nil {
		return fmt.Sprintf("%f", *sv.FloatP)
	}
	if sv.DurationP != nil {
		return sv.DurationP.String()
	}
	return ""
}

func (sv *SchemaValue) Set(text string) error {
	if sv.IntP != nil {
		i, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return err
		}
		sv.IntP = &i
		return nil
	}
	if sv.UintP != nil {
		i, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			return err
		}
		sv.UintP = &i
		return nil
	}
	if sv.StringP != nil {
		sv.StringP = &text
		return nil
	}
	if sv.FloatP != nil {
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		sv.FloatP = &f
		return nil
	}
	if sv.DurationP != nil {
		d, err := time.ParseDuration(text)
		if err != nil {
			return err
		}
		sv.DurationP = &d
		return nil
	}
	return nil
}

func (sv *SchemaValue) Type() string {
	if sv.IntP != nil {
		return "int64"
	}
	if sv.UintP != nil {
		return "uint64"
	}
	if sv.StringP != nil {
		return "string"
	}
	if sv.FloatP != nil {
		return "float64"
	}
	if sv.DurationP != nil {
		return "duration"
	}
	return "string"
}
