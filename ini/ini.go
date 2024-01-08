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

package ini

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path"
	"sort"
	"strings"
)

// Inventory contains parsed inventory representation
// Note: Groups and Hosts fields contain all the groups and hosts, not only top-level
type Inventory struct {
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

// ParseFile parses Inventory represented as a file
func ParseFile(f string) (*Inventory, error) {
	bs, err := os.ReadFile(f)
	if err != nil {
		return &Inventory{}, err
	}

	return Parse(bytes.NewReader(bs))
}

// ParseString parses Inventory represented as a string
func ParseString(input string) (*Inventory, error) {
	return Parse(strings.NewReader(input))
}

// Parse using some Reader
func Parse(r io.Reader) (*Inventory, error) {
	input := bufio.NewReader(r)
	inventory := &Inventory{}
	err := inventory.parse(input)
	if err != nil {
		return inventory, err
	}
	inventory.Reconcile()
	return inventory, nil
}

// Match looks for hosts that match the pattern
// Deprecated: Use `MatchHosts`, which does proper error handling
func (inventory *Inventory) Match(pattern string) []*Host {
	matchedHosts := make([]*Host, 0)
	for _, host := range inventory.Hosts {
		if m, err := path.Match(pattern, host.Name); err == nil && m {
			matchedHosts = append(matchedHosts, host)
		}
	}
	return matchedHosts
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
func (inventory *Inventory) HostsToLower() {
	inventory.Hosts = hostMapToLower(inventory.Hosts, false)
	for _, group := range inventory.Groups {
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
func (inventory *Inventory) GroupsToLower() {
	inventory.Groups = groupMapToLower(inventory.Groups, false)
	for _, host := range inventory.Hosts {
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
