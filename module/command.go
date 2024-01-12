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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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
