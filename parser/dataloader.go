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

// DataLoader-related helper methods

// Reconcile ensures inventory basic rules, run after updates.
// After initial inventory file processing, only direct relationships are set.
//
// This method:
//   - (re)sets Children and Parents for hosts and groups
//   - ensures that mandatory groups exist
//   - calculates variables for hosts and groups
func (dl *DataLoader) Reconcile() {
	// Clear all computed data
	for _, host := range dl.Hosts {
		host.clearData()
	}
	// a group can be empty (with no hosts in it), so the previous method will not clean it
	// on the other hand, a group could have been attached to a host by a user, but not added to the inventory.Groups map
	// so it's safer just to clean everything
	for _, group := range dl.Groups {
		group.clearData(make(map[string]struct{}, len(dl.Groups)))
	}

	allGroup := dl.getOrCreateGroup("all")
	ungroupedGroup := dl.getOrCreateGroup("ungrouped")
	ungroupedGroup.DirectParents[allGroup.Name] = allGroup

	// First, ensure that inventory.Groups contains all the groups
	for _, host := range dl.Hosts {
		for _, group := range host.DirectGroups {
			dl.Groups[group.Name] = group
			for _, ancestor := range group.ListParentGroupsOrdered() {
				dl.Groups[ancestor.Name] = ancestor
			}
		}
	}

	// Calculate intergroup relationships
	for _, group := range dl.Groups {
		group.DirectParents[allGroup.Name] = allGroup
		for _, ancestor := range group.ListParentGroupsOrdered() {
			group.Parents[ancestor.Name] = ancestor
			ancestor.Children[group.Name] = group
		}
	}

	// Now set hosts for groups and groups for hosts
	for _, host := range dl.Hosts {
		host.Groups[allGroup.Name] = allGroup
		for _, group := range host.DirectGroups {
			group.Hosts[host.Name] = host
			host.Groups[group.Name] = group
			for _, parent := range group.Parents {
				group.Parents[parent.Name] = parent
				parent.Children[group.Name] = group
				parent.Hosts[host.Name] = host
				host.Groups[parent.Name] = parent
			}
		}
	}
	dl.reconcileVars()
}

func (host *Host) clearData() {
	host.Groups = make(map[string]*Group)
	host.Vars = make(map[string]string)
	for _, group := range host.DirectGroups {
		group.clearData(make(map[string]struct{}, len(host.Groups)))
	}
}

func (group *Group) clearData(visited map[string]struct{}) {
	if _, ok := visited[group.Name]; ok {
		return
	}
	group.Hosts = make(map[string]*Host)
	group.Parents = make(map[string]*Group)
	group.Children = make(map[string]*Group)
	group.Vars = make(map[string]string)
	group.AllInventoryVars = nil
	group.AllFileVars = nil
	visited[group.Name] = struct{}{}
	for _, parent := range group.DirectParents {
		parent.clearData(visited)
	}
}

// getOrCreateGroup return group from inventory if exists or creates empty Group with given name
func (dl *DataLoader) getOrCreateGroup(groupName string) *Group {
	if group, ok := dl.Groups[groupName]; ok {
		return group
	}
	g := &Group{
		Name:     groupName,
		Hosts:    make(map[string]*Host),
		Vars:     make(map[string]string),
		Children: make(map[string]*Group),
		Parents:  make(map[string]*Group),

		DirectParents: make(map[string]*Group),
		InventoryVars: make(map[string]string),
		FileVars:      make(map[string]string),
	}
	dl.Groups[groupName] = g
	return g
}

// getOrCreateHost return host from inventory if exists or creates empty Host with given name
func (dl *DataLoader) getOrCreateHost(hostName string) *Host {
	if host, ok := dl.Hosts[hostName]; ok {
		return host
	}
	h := &Host{
		Name:   hostName,
		Port:   22,
		Groups: make(map[string]*Group),
		Vars:   make(map[string]string),

		DirectGroups:  make(map[string]*Group),
		InventoryVars: make(map[string]string),
		FileVars:      make(map[string]string),
	}
	dl.Hosts[hostName] = h
	return h
}

// addValues fills `to` map with values from `from` map
func addValues(to map[string]string, from map[string]string) {
	for k, v := range from {
		to[k] = v
	}
}

// copyStringMap creates a non-deep copy of the map
func copyStringMap(from map[string]string) map[string]string {
	result := make(map[string]string, len(from))
	addValues(result, from)
	return result
}
