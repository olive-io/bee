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

package stats

import (
	"encoding/json"
	"testing"
)

const result1 = `{
	"host": "127.0.0.1",
	"stdout": {
		"changed": true,
		"a": 1,
		"msg": "this is test message",
		"arr": [1, 2, 3],
		"sub": {
			"b": "a"
		}
	}
}`

func TestTaskStdout(t *testing.T) {
	r := &TaskResult{}
	err := json.Unmarshal([]byte(result1), r)
	if err != nil {
		t.Fatal(err)
	}
}
