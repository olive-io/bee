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

package parser

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const minMarshalDataLoader = `[Animals]
ET

[Animals:children]
Cats

[Cats]
Lion
`

//go:embed marshal_test_inventory.json
var minMarshalJSON string

func TestMarshalJSON(t *testing.T) {
	v := NewDataLoader()
	err := v.ParseString(minMarshalDataLoader)
	assert.Nil(t, err)

	j, err := json.MarshalIndent(v, "", "    ")
	assert.Nil(t, err)
	assert.Equal(t, minMarshalJSON, string(j))

	t.Run("unmarshal", func(t *testing.T) {
		var v2 DataLoader
		assert.Nil(t, json.Unmarshal(j, &v2))
		assert.Equal(t, v.Hosts["Lion"], v2.Hosts["Lion"])
		assert.Equal(t, v.Groups["Cats"], v2.Groups["Cats"])
	})
}

func TestMarshalWithVars(t *testing.T) {
	v := NewDataLoader()
	err := v.ParseFile("test_data/inventory")
	assert.Nil(t, err)

	v.HostsToLower()
	v.GroupsToLower()
	v.AddVarsLowerCased("test_data")

	j, err := json.MarshalIndent(v, "", "    ")
	assert.Nil(t, err)

	t.Run("unmarshal", func(t *testing.T) {
		var v2 DataLoader
		assert.Nil(t, json.Unmarshal(j, &v2))
		assert.Equal(t, v.Hosts["host1"], v2.Hosts["host1"])
		assert.Equal(t, v.Groups["tomcat"], v2.Groups["tomcat"])
	})
}
