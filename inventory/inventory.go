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

package inventory

import (
	"sync"

	"github.com/olive-io/bee/parser"
	"github.com/samber/lo"
)

type Manager struct {
	sync.RWMutex

	loader  *parser.DataLoader
	sources []string

	groups map[string]*parser.Group
	hosts  map[string]*parser.Host
}

func NewInventoryManager(loader *parser.DataLoader, sources ...string) (*Manager, error) {
	im := &Manager{
		loader:  loader,
		sources: sources,
		groups:  make(map[string]*parser.Group),
		hosts:   make(map[string]*parser.Host),
	}
	var err error
	im.groups, err = im.MatchedGroups()
	if err != nil {
		return nil, err
	}
	im.hosts, err = im.MatchedHosts()
	if err != nil {
		return nil, err
	}
	return im, nil
}

func (im *Manager) AddSources(sources ...string) error {
	im.Lock()
	defer im.Unlock()

	ins := make([]string, 0)
	for _, in := range sources {
		_, ok := lo.Find[string](im.sources, func(item string) bool {
			return item == in
		})
		if !ok {
			ins = append(ins, in)
		}
	}

	im.sources = append(im.sources, ins...)
	for _, source := range ins {
		matched, err := im.loader.MatchGroups(source)
		if err != nil {
			return err
		}
		for key, host := range matched {
			im.groups[key] = host
		}
	}

	for _, source := range ins {
		matched, err := im.loader.MatchHosts(source)
		if err != nil {
			return err
		}
		for key, host := range matched {
			im.hosts[key] = host
		}
	}
	return nil
}

func (im *Manager) MatchedGroups() (map[string]*parser.Group, error) {
	im.RLock()
	defer im.RUnlock()
	groups := make(map[string]*parser.Group)
	for _, source := range im.sources {
		matched, err := im.loader.MatchGroups(source)
		if err != nil {
			return nil, err
		}
		for key, host := range matched {
			groups[key] = host
		}
	}
	return groups, nil
}

func (im *Manager) MatchedHosts() (map[string]*parser.Host, error) {
	im.RLock()
	defer im.RUnlock()
	hosts := make(map[string]*parser.Host)
	for _, source := range im.sources {
		matched, err := im.loader.MatchHosts(source)
		if err != nil {
			return nil, err
		}
		for key, host := range matched {
			hosts[key] = host
		}
	}
	return hosts, nil
}

func (im *Manager) FindGroup(name string) (*parser.Group, bool) {
	group, ok := im.groups[name]
	return group, ok
}

func (im *Manager) FindHost(name string) (*parser.Host, bool) {
	host, ok := im.hosts[name]
	return host, ok
}
