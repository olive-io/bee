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

package copy

import (
	"path"
	"strings"

	"github.com/olive-io/bee/executor/client"
	"github.com/olive-io/bee/module"
	"github.com/olive-io/bee/vars"
)

const copyExample = ``

var CopyModule = &module.Module{
	Command: &module.Command{
		Name:    "bee.builtin.copy",
		Long:    "",
		Script:  "builtin/copy/copy.tengo",
		Authors: []string{"lack"},
		Version: "v1.0.0",
		Example: copyExample,
		Params: []*module.Schema{
			{
				Name:        "src",
				Type:        "string",
				Description: "",
			},
			{
				Name:        "dst",
				Type:        "string",
				Description: "",
			},
		},
		Returns: []*module.Schema{},
		PreRun:  preRun,
		Run:     module.DefaultRunCommand,
	},
	Dir: "builtin/copy",
}

var preRun module.RunE = func(ctx *module.RunContext, options ...client.ExecOption) ([]byte, error) {
	out := []byte("")

	fs := ctx.Cmd.Flags()
	src, err := fs.GetString("src")
	if err != nil {
		return nil, err
	}
	dst, err := fs.GetString("dst")
	if err != nil {
		return nil, err
	}

	home := ctx.Variables.GetDefault(vars.BeeHome, ".bee")
	goos := ctx.Variables.GetDefault(vars.BeePlatformVars, "linux")

	dst = path.Join(home, "tmp", path.Base(dst))
	if goos == "windows" {
		dst = strings.ReplaceAll(dst, "/", "\\")
	}

	err = ctx.Conn.Put(ctx, src, dst)
	if err != nil {
		return nil, err
	}

	ctx.Variables.Set(module.PrefixFlag+"src", dst)

	return out, nil
}
