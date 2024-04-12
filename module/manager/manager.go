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

package manager

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"go.uber.org/zap"

	"github.com/olive-io/bee/module"
	"github.com/olive-io/bee/module/hook"
)

type Manager struct {
	dir string

	mu      sync.RWMutex
	modules map[string]*module.Module

	lg *zap.Logger
}

func NewModuleManager(lg *zap.Logger, dir string) (*Manager, error) {
	var err error
	rdir := filepath.Join(dir, "repl")
	if _, err = os.Stat(rdir); err != nil {
		return nil, err
	}

	mdir := filepath.Join(dir, "modules")
	mg := &Manager{
		dir:     mdir,
		modules: map[string]*module.Module{},
		lg:      lg,
	}
	if err = mg.LoadDir(mdir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	return mg, nil
}

func (mg *Manager) RootDir() string {
	return mg.dir
}

func (mg *Manager) Modules() []*module.Module {
	mg.mu.RLock()
	defer mg.mu.RUnlock()
	ms := make([]*module.Module, 0)
	for name, _ := range mg.modules {
		ms = append(ms, mg.modules[name])
	}
	return ms
}

func (mg *Manager) LoadModule(m *module.Module) {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	mg.modules[m.Name] = m
}

// LoadDir loads modules from local directory
func (mg *Manager) LoadDir(dir string) error {
	if mg.validDir(dir) {
		m, err := module.LoadDir(dir)
		if err != nil {
			return err
		}
		if m.Hide {
			return nil
		}
		if hk, ok := hook.Hooks[m.Name]; ok {
			forCommandHook(&m.Command, hk)
		}
		for i := range m.Commands {
			cmd := m.Commands[i]
			name := m.Name + cmd.Name
			if hk, ok := hook.Hooks[name]; ok {
				forCommandHook(&m.Command, hk)
			}
		}
		mg.LoadModule(m)
		return nil
	}

	ents, _ := os.ReadDir(dir)
	dirs := make([]string, 0)
	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}
		dirs = append(dirs, filepath.Join(dir, ent.Name()))
	}
	for _, sd := range dirs {
		err := mg.LoadDir(sd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mg *Manager) validDir(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		return false
	}

	stat, _ := os.Stat(filepath.Join(dir, "bee.yml"))
	if stat != nil {
		return true
	}
	stat, _ = os.Stat(filepath.Join(dir, "bee.yaml"))
	if stat != nil {
		return true
	}
	return false
}

// Find returns the *Module by name
func (mg *Manager) Find(name string) (*module.Module, bool) {
	mg.mu.RLock()
	m, ok := mg.modules[name]
	mg.mu.RUnlock()
	if !ok {
		if !strings.Contains(name, ".") {
			name = "bee.builtin." + name
			return mg.Find(name)
		}
		return nil, false
	}
	return m, ok
}

func forCommandHook(cmd *module.Command, hk *hook.CommandHook) {
	if hk.PreRun != nil {
		cmd.PreRun = hk.PreRun
	}
	if hk.Run != nil {
		cmd.Run = hk.Run
	}
	if hk.PostRun != nil {
		cmd.PostRun = hk.PostRun
	}
}
