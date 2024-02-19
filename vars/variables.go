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

package vars

import (
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/parser"
)

type VariableManager struct {
	loader    *parser.DataLoader
	inventory *inv.Manager

	groupVariables map[string]map[string]string
	hostVariables  map[string]map[string]string
}

func NewVariablesManager(loader *parser.DataLoader, inventory *inv.Manager) *VariableManager {
	vm := &VariableManager{
		loader:    loader,
		inventory: inventory,
	}

	vm.Reconcile()
	return vm
}

func (vm *VariableManager) Reconcile() {
	groupVariables := map[string]map[string]string{}
	groups, _ := vm.loader.MatchGroups("*")
	for name, group := range groups {
		groupVariables[name] = group.Vars
	}

	hosts, _ := vm.loader.MatchHosts("*")
	hostVariables := map[string]map[string]string{}
	for name, host := range hosts {
		hostVariables[name] = host.Vars
	}

	vm.groupVariables = groupVariables
	vm.hostVariables = hostVariables
}

func (vm *VariableManager) MustGetHostDefaultValue(host, name, defaultV string) string {
	hv, ok := vm.hostVariables[host]
	if !ok {
		return defaultV
	}
	value, ok := hv[name]
	if !ok {
		return defaultV
	}
	return value
}
