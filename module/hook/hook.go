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

import "github.com/olive-io/bee/module"

var Hooks = map[string]*CommandHook{
	"bee.builtin.copy":  copyHook,
	"bee.builtin.fetch": fetchHook,
}

type CommandHook struct {
	PreRun  module.RunE
	Run     module.RunE
	PostRun module.RunE
}
