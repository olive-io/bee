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

package filter

type IFilter interface {
	OnPreTaskProps(pros, headers map[string]any) (map[string]any, map[string]any)
	OnPostTaskStdout(stdout map[string]any) map[string]any
}

func NewFilter() IFilter {
	return &BaseFilter{}
}

type BaseFilter struct{}

func (b *BaseFilter) OnPreTaskProps(pros, headers map[string]any) (map[string]any, map[string]any) {
	return pros, headers
}

func (b *BaseFilter) OnPostTaskStdout(stdout map[string]any) map[string]any {
	return stdout
}