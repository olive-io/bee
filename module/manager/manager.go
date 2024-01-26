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

package manager

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/olive-io/bee/module"
	"github.com/olive-io/bee/module/hook"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type Manager struct {
	dir        string
	moduleDirs []string
	modules    map[string]*module.Module

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
		dir:        mdir,
		moduleDirs: []string{},
		modules:    map[string]*module.Module{},
		lg:         lg,
	}
	if err = mg.LoadDir(mdir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		mg.moduleDirs = append(mg.moduleDirs, mdir)
	}

	return mg, nil
}

func (mg *Manager) RootDir() string {
	return mg.dir
}

func (mg *Manager) ModuleDirs() []string {
	dirs := make([]string, len(mg.moduleDirs))
	for i, dir := range mg.moduleDirs {
		dirs[i] = dir
	}
	return dirs
}

func (mg *Manager) LoadModule(m *module.Module) {
	mg.modules[m.Name] = m
}

// LoadDir loads modules from local directory
func (mg *Manager) LoadDir(dir string) error {
	if mg.validDir(dir) {
		mg.moduleDirs = lo.Uniq[string](append(mg.moduleDirs, dir))
		m, err := module.LoadDir(dir)
		if err != nil {
			return err
		}
		if hk, ok := hook.Hooks[m.Name]; ok {
			forCommandHook(m.Command, hk)
		}
		for i := range m.Commands {
			cmd := m.Commands[i]
			name := m.Name + cmd.Name
			if hk, ok := hook.Hooks[name]; ok {
				forCommandHook(m.Command, hk)
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
	m, ok := mg.modules[name]
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
