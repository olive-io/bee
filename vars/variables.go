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
