/*
   Copyright 2024 The bee Authors

   This library is free software; you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public
   License as published by the Free Software Foundation; either
   version 2.1 of the License, or (at your option) any later version.

   This library is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License along with this library;
*/

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
