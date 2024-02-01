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
	"encoding/json"

	"github.com/samber/lo"
	"golang.org/x/exp/maps"
)

type alwaysNil interface{} // to hold place for Group and Host references; must be nil in serialized form

func (group *Group) MarshalJSON() ([]byte, error) {
	type groupWithoutCustomMarshal Group

	return json.Marshal(&struct {
		groupWithoutCustomMarshal
		Hosts         map[string]alwaysNil
		Children      map[string]alwaysNil
		Parents       map[string]alwaysNil
		DirectParents map[string]alwaysNil
	}{
		groupWithoutCustomMarshal: groupWithoutCustomMarshal(*group),
		Hosts:                     makeNilValueMap(group.Hosts),
		Children:                  makeNilValueMap(group.Children),
		Parents:                   makeNilValueMap(group.Parents),
		DirectParents:             makeNilValueMap(group.DirectParents),
	})
}

func (host *Host) MarshalJSON() ([]byte, error) {
	type hostWithoutCustomMarshal Host

	return json.Marshal(&struct {
		hostWithoutCustomMarshal
		Groups       map[string]alwaysNil
		DirectGroups map[string]alwaysNil
	}{
		hostWithoutCustomMarshal: hostWithoutCustomMarshal(*host),
		Groups:                   makeNilValueMap(host.Groups),
		DirectGroups:             makeNilValueMap(host.DirectGroups),
	})
}

func makeNilValueMap[K comparable, V any](m map[K]*V) map[K]alwaysNil {
	return lo.MapValues(m, func(_ *V, _ K) alwaysNil { return nil })
}

func (dl *DataLoader) UnmarshalJSON(data []byte) error {
	type inventoryWithoutCustomUnmarshal DataLoader
	var rawDataLoader inventoryWithoutCustomUnmarshal
	if err := json.Unmarshal(data, &rawDataLoader); err != nil {
		return err
	}
	// rawDataLoader's Groups and Hosts should now contain all properties,
	// except child group maps and host maps are filled with original keys and null values

	// reassign child groups and hosts to reference rawDataLoader.Hosts and .Groups

	for _, group := range rawDataLoader.Groups {
		group.Hosts = lo.PickByKeys(rawDataLoader.Hosts, maps.Keys(group.Hosts))
		group.Children = lo.PickByKeys(rawDataLoader.Groups, maps.Keys(group.Children))
		group.Parents = lo.PickByKeys(rawDataLoader.Groups, maps.Keys(group.Parents))
		group.DirectParents = lo.PickByKeys(rawDataLoader.Groups, maps.Keys(group.DirectParents))
	}

	for _, host := range rawDataLoader.Hosts {
		host.Groups = lo.PickByKeys(rawDataLoader.Groups, maps.Keys(host.Groups))
		host.DirectGroups = lo.PickByKeys(rawDataLoader.Groups, maps.Keys(host.DirectGroups))
	}

	dl.mu.Lock()
	defer dl.mu.Unlock()

	dl.Groups = rawDataLoader.Groups
	dl.Hosts = rawDataLoader.Hosts
	return nil
}
