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

package hook

import (
	"github.com/olive-io/bee/executor/client"
	"github.com/olive-io/bee/module"
)

var fetchHook = &CommandHook{
	PreRun:  fetchPreRun,
	PostRun: fetchPostRun,
}

var fetchPreRun module.RunE = func(ctx *module.RunContext, options ...client.ExecOption) ([]byte, error) {
	return []byte(""), nil
}

var fetchPostRun module.RunE = func(ctx *module.RunContext, options ...client.ExecOption) ([]byte, error) {
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

	err = ctx.Conn.Get(ctx, src, dst)
	if err != nil {
		return nil, err
	}

	return out, nil
}
