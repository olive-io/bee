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
