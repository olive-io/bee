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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAncestorsOrdered(t *testing.T) {
	v := parseString(t, `
	host1
	[notMyGroup3]
	[myGroup2]
	[myGroup1]
	host1
	[myGroup2:children]
	myGroup1
	[notMyGroup3:children]
	myGroup2
	`)

	host1 := v.Hosts["host1"]
	assert.NotNil(t, host1)
	assert.Len(t, host1.Groups, 4)

	groups := host1.ListGroupsOrdered()
	assert.Len(t, groups, 4)
	assert.Equal(t, groups[0].Name, "myGroup1")
	assert.Equal(t, groups[1].Name, "myGroup2")
	assert.Equal(t, groups[2].Name, "notMyGroup3")
	assert.Equal(t, groups[3].Name, "all")

	group1 := v.Groups["myGroup1"]
	assert.NotNil(t, group1)
	groups = group1.ListParentGroupsOrdered()
	assert.NotNil(t, groups)

	assert.Len(t, groups, 3)
	assert.Equal(t, groups[0].Name, "myGroup2")
	assert.Equal(t, groups[1].Name, "notMyGroup3")
	assert.Equal(t, groups[2].Name, "all")
}

func TestMatchGroupsOrdered(t *testing.T) {
	v := parseString(t, `
	host1
	[notMyGroup3]
	[myGroup2]
	[myGroup1]
	host1
	[myGroup2:children]
	myGroup1
	[notMyGroup3:children]
	myGroup2
	`)

	host1 := v.Hosts["host1"]
	assert.NotNil(t, host1)
	assert.Len(t, host1.Groups, 4)

	groups, err := host1.MatchGroupsOrdered("my*")
	assert.Nil(t, err)
	assert.Len(t, groups, 2)
	assert.Equal(t, groups[0].Name, "myGroup1")
	assert.Equal(t, groups[1].Name, "myGroup2")

	group1 := v.Groups["myGroup1"]
	assert.NotNil(t, group1)
	groups, err = group1.MatchGroupsOrdered("my*")
	assert.Nil(t, err)
	assert.NotNil(t, groups)

	assert.Len(t, groups, 1)
	assert.Equal(t, groups[0].Name, "myGroup2")
}
