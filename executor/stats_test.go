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

package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregateStats(t *testing.T) {
	as := NewStats()

	as.SetCustomStats("xx", map[string]string{"a": "1"}, "host1")
	m := map[string]string{"b": "3", "a": "1"}
	as.UpdateCustomStats("xx", m, "host1")

	value, ok := as.GetCustomStats("host1", "xx")
	assert.Equal(t, true, ok)
	sm, ok := value.(map[string]string)
	assert.Equal(t, true, ok)
	assert.Equal(t, "1", sm["a"])

	as.UpdateCustomStats("i", 2, "host1")
	as.UpdateCustomStats("i", 3, "host1")
	value, _ = as.GetCustomStats("host1", "i")
	assert.Equal(t, int64(5), value)

	as.UpdateCustomStats("f", 2.1, "host1")
	as.UpdateCustomStats("f", 3.2, "host1")
	value, _ = as.GetCustomStats("host1", "f")
	assert.Equal(t, int64(5), int64(value.(float64)))

	as.UpdateCustomStats("s", "hello", "host1")
	as.UpdateCustomStats("s", " ", "host1")
	as.UpdateCustomStats("s", "world", "host1")

	value, _ = as.GetCustomStats("host1", "s")
	assert.Equal(t, "hello world", value)
}
