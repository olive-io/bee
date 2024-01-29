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
	"path"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/olive-io/bee/executor/client"
	"github.com/olive-io/bee/vars"
)

const (
	PrefixFlag = "__flag_"
)

type StableMap struct {
	store map[string]string
}

func NewVariables() *StableMap {
	return &StableMap{store: map[string]string{}}
}

func (sm *StableMap) Set(key string, value string) {
	sm.store[key] = value
}

func (sm *StableMap) GetDefault(key string, defaultV string) string {
	v, ok := sm.store[key]
	if !ok {
		return defaultV
	}
	return v
}

func (sm *StableMap) Exists(key string) bool {
	_, ok := sm.store[key]
	return ok
}

type RunContext struct {
	context.Context

	Logger    *zap.Logger
	Cmd       *Command
	Conn      client.IClient
	Variables *StableMap
}

type RunE func(ctx *RunContext, options ...client.ExecOption) ([]byte, error)

type Command struct {
	Name     string         `yaml:"name"`
	Long     string         `yaml:"long"`
	Script   string         `yaml:"script"`
	Authors  []string       `yaml:"authors"`
	Version  string         `yaml:"version"`
	Example  string         `yaml:"example"`
	Params   []*Schema      `yaml:"params"`
	Returns  []*Schema      `yaml:"returns"`
	Commands []*Command     `yaml:"commands"`
	Root     string         `yaml:"root"`
	cobra    *cobra.Command `yaml:"-"`

	PreRun  RunE `yaml:"-"`
	Run     RunE `yaml:"-"`
	PostRun RunE `yaml:"-"`
}

func (c *Command) NewContext(ctx context.Context, lg *zap.Logger, conn client.IClient, variables *StableMap) *RunContext {
	rctx := &RunContext{
		Context:   ctx,
		Logger:    lg,
		Cmd:       c,
		Conn:      conn,
		Variables: variables,
	}
	return rctx
}

func (c *Command) Runnable() bool {
	return c.Script != ""
}

func (c *Command) ParseCmd() *cobra.Command {
	if c.cobra != nil {
		return c.cobra
	}

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

var DefaultRunCommand RunE = func(ctx *RunContext, opts ...client.ExecOption) ([]byte, error) {
	command := ctx.Cmd
	lg := ctx.Logger
	conn := ctx.Conn
	execOptions := client.NewExecOptions()
	for _, opt := range opts {
		opt(execOptions)
	}

	command.Flags().VisitAll(func(flag *pflag.Flag) {
		value := ctx.Variables.GetDefault(PrefixFlag+flag.Name, flag.Value.String())
		arg := "--" + flag.Name + "=" + value
		execOptions.Args = append(execOptions.Args, arg)
	})

	options := make([]client.ExecOption, 0)
	ext, ok := KnownExt(path.Ext(command.Script))
	if !ok {
		ext = Tengo
	}

	home := ctx.Variables.GetDefault(vars.BeeHome, ".bee")
	goos := ctx.Variables.GetDefault(vars.BeePlatformVars, "linux")

	var repl string
	var err error
	if repl, err = checkRepl(goos, ext); err != nil {
		return nil, err
	}

	script := path.Join(home, "modules", command.Root, command.Script)
	if goos == "windows" {
		script = strings.ReplaceAll(script, "/", "\\")
	}

	switch ext {
	case Tengo:
		repl = path.Join(home, "bin", repl)
		if goos == "windows" {
			repl = strings.ReplaceAll(repl, "/", "\\")
		}
	case Bash:
		options = append(options, client.ExecWithArgs("-c"))
	case Powershell:
	}

	options = append(options, client.ExecWithArgs(script))
	options = append(options, client.ExecWithArgs(execOptions.Args...))
	for key, value := range execOptions.Environments {
		options = append(options, client.ExecWithEnv(key, value))
	}

	shell := fmt.Sprintf("%s %s %s", repl, script, strings.Join(execOptions.Args, " "))
	lg.Debug("remote execute", zap.String("command", shell))
	cmd, err := conn.Execute(ctx, repl, options...)
	if err != nil {
		return nil, err
	}
	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, string(beautify(data)))
	}
	return beautify(data), nil
}
