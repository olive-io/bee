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

package ping

import (
	"github.com/olive-io/bee/module"
)

const pingExample = ``

var PingModule = &module.Module{
	Command: &module.Command{
		Name:    "bee.builtin.ping",
		Long:    "",
		Script:  "builtin/ping/ping.tengo",
		Authors: []string{"lack"},
		Version: "v1.0.0",
		Example: pingExample,
		Params: []*module.Schema{
			{
				Name:        "data",
				Type:        "string",
				Description: "Data to return for the `ping` return value.If this parameter is set to `crash`, the module will cause an error.",
				Default:     "ping",
			},
		},
		Returns: []*module.Schema{
			{
				Name:    "data",
				Type:    "string",
				Default: "pong",
			},
		},

		Run: module.DefaultRunCommand,
	},
	Dir: "builtin/ping",
}
