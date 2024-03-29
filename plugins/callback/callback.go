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

package callback

import "github.com/olive-io/bee/stats"

type ICallBack interface {
	RunnerOnUnreachable(result *stats.TaskResult)
	RunnerOnOk(result *stats.TaskResult)
	RunnerOkFailed(result *stats.TaskResult)
}

func NewCallBack() ICallBack {
	return &BaseCallBack{}
}

type BaseCallBack struct {
}

func (b *BaseCallBack) RunnerOnUnreachable(result *stats.TaskResult) {
}

func (b *BaseCallBack) RunnerOnOk(result *stats.TaskResult) {
}

func (b *BaseCallBack) RunnerOkFailed(result *stats.TaskResult) {
}
