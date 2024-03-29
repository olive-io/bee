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

package extra

import (
	"github.com/d5/tengo/v2"
)

type BoolP struct {
	tengo.ObjectImpl

	Value bool
}

func (b *BoolP) TypeName() string {
	return "bool"
}

func (b *BoolP) Copy() tengo.Object {
	return &BoolP{Value: b.Value}
}

func (b *BoolP) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

func (b *BoolP) Set(text string) error {
	if text == "true" || text == "1" {
		b.Value = true
	} else {
		b.Value = false
	}
	return nil
}

func (b *BoolP) Type() string {
	return "bool"
}
