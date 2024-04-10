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
	"bufio"
	"bytes"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
)

// DataLoader contains parsed inventory representation
// Note: Groups and Hosts fields contain all the groups and hosts, not only top-level
type DataLoader struct {
	mu sync.RWMutex

	Groups map[string]*Group
	Hosts  map[string]*Host
}

// Group represents ansible group
type Group struct {
	Name     string
	Vars     map[string]string
	Hosts    map[string]*Host
	Children map[string]*Group
	Parents  map[string]*Group

	DirectParents map[string]*Group
	// Vars set in inventory
	InventoryVars map[string]string
	// Vars set in group_vars
	FileVars map[string]string
	// Projection of all parent inventory variables
	AllInventoryVars map[string]string
	// Projection of all parent group_vars variables
	AllFileVars map[string]string
}

// Host represents ansible host
type Host struct {
	Name   string
	Port   int
	Vars   map[string]string
	Groups map[string]*Group

	DirectGroups map[string]*Group
	// Vars set in inventory
	InventoryVars map[string]string
	// Vars set in host_vars
	FileVars map[string]string
}

func NewDataLoader() *DataLoader {
	dl := &DataLoader{}
	dl.initData()
	return dl
}

// ParseFile parses DataLoader represented as a file
func (dl *DataLoader) ParseFile(f string) error {
	bs, err := os.ReadFile(f)
	if err != nil {
		return err
	}

	return dl.Parse(bytes.NewReader(bs))
}

// ParseString parses DataLoader represented as a string
func (dl *DataLoader) ParseString(input string) error {
	return dl.Parse(strings.NewReader(input))
}

// Parse using some Reader
func (dl *DataLoader) Parse(r io.Reader) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	input := bufio.NewReader(r)
	err := dl.parse(input)
	if err != nil {
		return err
	}
	dl.Reconcile()
	return nil
}

// GroupMapListValues transforms map of Groups into Group list in lexical order
func GroupMapListValues(mymap map[string]*Group) []*Group {
	values := make([]*Group, len(mymap))

	i := 0
	for _, v := range mymap {
		values[i] = v
		i++
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].Name < values[j].Name
	})
	return values
}

// HostMapListValues transforms map of Hosts into Host list in lexical order
func HostMapListValues(mymap map[string]*Host) []*Host {
	values := make([]*Host, len(mymap))

	i := 0
	for _, v := range mymap {
		values[i] = v
		i++
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].Name < values[j].Name
	})
	return values
}

// HostsToLower transforms all host names to lowercase
func (dl *DataLoader) HostsToLower() {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.Hosts = hostMapToLower(dl.Hosts, false)
	for _, group := range dl.Groups {
		group.Hosts = hostMapToLower(group.Hosts, true)
	}
}

func hostMapToLower(hosts map[string]*Host, keysOnly bool) map[string]*Host {
	newHosts := make(map[string]*Host, len(hosts))
	for hostname, host := range hosts {
		hostname = strings.ToLower(hostname)
		if !keysOnly {
			host.Name = hostname
		}
		newHosts[hostname] = host
	}
	return newHosts
}

// GroupsToLower transforms all group names to lowercase
func (dl *DataLoader) GroupsToLower() {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.Groups = groupMapToLower(dl.Groups, false)
	for _, host := range dl.Hosts {
		host.DirectGroups = groupMapToLower(host.DirectGroups, true)
		host.Groups = groupMapToLower(host.Groups, true)
	}
}

func (group *Group) String() string {
	return group.Name
}

func (host *Host) String() string {
	return host.Name
}

func groupMapToLower(groups map[string]*Group, keysOnly bool) map[string]*Group {
	newGroups := make(map[string]*Group, len(groups))
	for groupname, group := range groups {
		groupname = strings.ToLower(groupname)
		if !keysOnly {
			group.Name = groupname
			group.DirectParents = groupMapToLower(group.DirectParents, true)
			group.Parents = groupMapToLower(group.Parents, true)
			group.Children = groupMapToLower(group.Children, true)
		}
		newGroups[groupname] = group
	}
	return newGroups
}
