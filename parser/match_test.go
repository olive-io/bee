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

func TestGroupsMatching(t *testing.T) {
	v := parseString(t, `
	host1
	host2
	[myGroup1]
	host1
	[myGroup2]
	host1
	[groupForCats]
	host1
	`)

	groups, err := v.MatchGroups("*Group*")
	assert.Nil(t, err)
	assert.Contains(t, groups, "myGroup1")
	assert.Contains(t, groups, "myGroup2")
	assert.Len(t, groups, 2)

	groups, err = v.Hosts["host1"].MatchGroups("*Group*")
	assert.Nil(t, err)
	assert.Contains(t, groups, "myGroup1")
	assert.Contains(t, groups, "myGroup2")
	assert.Len(t, groups, 2)
}

func TestHostsMatching(t *testing.T) {
	v := parseString(t, `
	myHost1
	otherHost2
	[group1]
	myHost1
	[group2]
	myHost1
	myHost2
	`)

	hosts, err := v.MatchHosts("my*")
	assert.Nil(t, err)
	assert.Contains(t, hosts, "myHost1")
	assert.Contains(t, hosts, "myHost2")
	assert.Len(t, hosts, 2)

	hosts, err = v.Groups["group1"].MatchHosts("*my*")
	assert.Nil(t, err)
	assert.Contains(t, hosts, "myHost1")
	assert.Len(t, hosts, 1)

	hosts, err = v.Groups["group2"].MatchHosts("*my*")
	assert.Nil(t, err)
	assert.Contains(t, hosts, "myHost1")
	assert.Contains(t, hosts, "myHost2")
	assert.Len(t, hosts, 2)
}

func TestVarsMatching(t *testing.T) {
	v := parseString(t, `
	host1 myHostVar=myHostVarValue otherHostVar=otherHostVarValue
	
	[group1]
	host1

	[group1:vars]
	myGroupVar=myGroupVarValue
	otherGroupVar=otherGroupVarValue
	`)
	group := v.Groups["group1"]
	vars, err := group.MatchVars("my*")
	assert.Nil(t, err)
	assert.Contains(t, vars, "myGroupVar")
	assert.Len(t, vars, 1)
	assert.Equal(t, "myGroupVarValue", vars["myGroupVar"])

	host := v.Hosts["host1"]
	vars, err = host.MatchVars("my*")
	assert.Nil(t, err)
	assert.Contains(t, vars, "myHostVar")
	assert.Contains(t, vars, "myGroupVar")
	assert.Len(t, vars, 2)
	assert.Equal(t, "myHostVarValue", vars["myHostVar"])
	assert.Equal(t, "myGroupVarValue", vars["myGroupVar"])
}
