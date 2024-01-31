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

package bee

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	bexecutor "github.com/olive-io/bee/executor"
	"github.com/olive-io/bee/executor/client"
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/module"
	mmg "github.com/olive-io/bee/module/manager"
	"github.com/olive-io/bee/parser"
	"github.com/olive-io/bee/secret"
	"github.com/olive-io/bee/vars"
)

const (
	syncFlag = "sync"
)

type Runtime struct {
	opts *Options

	pool *ants.Pool

	db *pebble.DB

	inventory *inv.Manager
	variables *vars.VariableManager
	loader    *parser.DataLoader
	passwords *secret.PasswordManager
	modules   *mmg.Manager
	executor  *bexecutor.Executor
}

func NewRuntime(
	inventory *inv.Manager, variables *vars.VariableManager,
	loader *parser.DataLoader, opts ...Option,
) (*Runtime, error) {

	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	lg := options.logger
	antLogger := &antsLogger{lg: lg}
	poolSize := options.parallel
	antOpts := ants.Options{
		PreAlloc:         true,
		MaxBlockingTasks: poolSize,
		Logger:           antLogger,
	}
	pool, err := ants.NewPool(poolSize, ants.WithOptions(antOpts))
	if err != nil {
		return nil, err
	}

	dbDir := filepath.Join(options.dir, "db")
	db, err := openDB(lg, dbDir)
	if err != nil {
		return nil, errors.Wrapf(err, "open embded db")
	}

	passwords := secret.NewPasswordManager(lg, db)
	executor := bexecutor.NewExecutor(lg, inventory, passwords)
	modules, err := mmg.NewModuleManager(lg, options.dir)
	if err != nil {
		return nil, err
	}

	rt := &Runtime{
		opts:      options,
		pool:      pool,
		db:        db,
		inventory: inventory,
		variables: variables,
		loader:    loader,
		passwords: passwords,
		executor:  executor,
		modules:   modules,
	}

	if err = rt.loadModules(); err != nil {
		return nil, err
	}

	return rt, nil
}

func (rt *Runtime) loadModules() error {
	multiPath := rt.opts.modulePath
	for _, dir := range multiPath {
		err := rt.modules.LoadDir(dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rt *Runtime) Logger() *zap.Logger {
	return rt.opts.logger
}

func (rt *Runtime) Inventory() *inv.Manager {
	return rt.inventory
}

func (rt *Runtime) Execute(ctx context.Context, host, shell string, opts ...RunOption) ([]byte, error) {
	ech := make(chan error, 1)
	ch := make(chan []byte, 1)
	defer func() {
		close(ch)
		close(ech)
	}()

	err := rt.pool.Submit(func() {
		data, err := rt.run(ctx, host, shell, opts...)
		if err != nil {
			ech <- err
			return
		}
		ch <- data
	})

	if err != nil {
		return nil, err
	}

	select {
	case err = <-ech:
		return nil, err
	case data := <-ch:
		return data, nil
	}
}

func (rt *Runtime) run(ctx context.Context, host, shell string, opts ...RunOption) ([]byte, error) {
	lg := rt.Logger()
	options := newRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	args := strings.Split(shell, " ")
	mname := args[0]
	if len(args) > 1 {
		args = args[1:]
	}

	if before, after, ok := strings.Cut(mname, "."); ok {
		mname = before
		args = append(args, after)
	}
	bm, ok := rt.modules.Find(mname)
	if !ok {
		return nil, errors.New("unknown module")
	}

	cmd, err := bm.Execute(args...)
	if err != nil {
		return nil, err
	}

	if !cmd.Runnable() {
		return nil, errors.New("command can't be execute")
	}

	conn, err := rt.executor.GetClient(host)
	if err != nil {
		return nil, err
	}

	sm := rt.applyStableMap(host)
	if options.sync {
		sm.Set(syncFlag, "")
	}

	if err = rt.syncRepl(ctx, conn, sm); err != nil {
		return nil, err
	}

	if err = rt.syncDepModules(ctx, conn, sm); err != nil {
		return nil, err
	}

	if err = rt.syncModule(ctx, conn, bm, sm); err != nil {
		return nil, err
	}

	rctx := cmd.NewContext(ctx, lg, conn, sm)
	execOptions := make([]client.ExecOption, 0)

	if cmd.PreRun != nil {
		if _, err = cmd.PreRun(rctx, execOptions...); err != nil {
			lg.Error("execute prepare command", zap.Error(err))
		}
	}
	if cmd.Run == nil {
		return nil, errors.New("command can not be executed")
	}
	out, err := cmd.Run(rctx, execOptions...)
	if err != nil {
		return nil, err
	}
	if cmd.PostRun != nil {
		if _, err = cmd.PostRun(rctx, execOptions...); err != nil {
			lg.Error("execute post command", zap.Error(err))
		}
	}
	return out, err
}

func (rt *Runtime) syncRepl(ctx context.Context, conn client.IClient, sm *module.StableMap) error {
	lg := rt.opts.logger
	home := sm.GetDefault(vars.BeeHome, ".bee")
	goos := sm.GetDefault(vars.BeePlatformVars, "linux")
	arch := sm.GetDefault(vars.BeeArchVars, "amd64")

	repl := path.Join(home, "bin", "tengo")
	if goos == "windows" {
		repl = strings.ReplaceAll(repl, "/", "\\")
		repl += ".exe"
	}

	toSync := sm.Exists(syncFlag)
	if !toSync {
		cmd, err := conn.Execute(ctx, repl, client.ExecWithArgs("-version"))
		if err != nil {
			return err
		}
		_, err = cmd.CombinedOutput()
		if err != nil {
			toSync = true
		}
	}

	if !toSync {
		return nil
	}

	toolchain := filepath.Join(rt.opts.dir, "repl", "tengo."+goos+"."+arch)
	if goos == "windows" {
		toolchain += ".exe"
	}
	_, err := os.Stat(toolchain)
	if err != nil {
		return err
	}

	lg.Debug("upload toolchain",
		zap.String("repl", "tengo"),
		zap.String("platform", goos),
		zap.String("arch", arch),
		zap.String("remote", repl),
	)

	return conn.Put(ctx, toolchain, repl, client.PutWithMkdir(true))
}

func (rt *Runtime) syncDepModules(ctx context.Context, conn client.IClient, sm *module.StableMap) error {
	root := rt.modules.RootDir()
	modules := rt.modules.Modules()

	lg := rt.Logger()
	home := sm.GetDefault(vars.BeeHome, ".bee")
	goos := sm.GetDefault(vars.BeePlatformVars, "linux")

	for _, item := range modules {
		if !strings.HasPrefix(item.Name, "bee.builtin") {
			continue
		}
		localDir := filepath.Join(root, item.Root)
		remoteDir := path.Join(home, "modules", item.Root)
		if goos == "windows" {
			remoteDir = strings.ReplaceAll(remoteDir, "/", "\\")
		}
		rs, _ := conn.Stat(ctx, remoteDir)
		if rs != nil {
			continue
		}
		lg.Debug("put bee module",
			zap.String("name", item.Name),
			zap.String("local", localDir),
			zap.String("remote", remoteDir))
		err := conn.Put(ctx, localDir, remoteDir, client.PutWithDir(true))
		if err != nil {
			return err
		}
	}

	return nil
}

func (rt *Runtime) syncModule(ctx context.Context, conn client.IClient, bm *module.Module, sm *module.StableMap) error {
	root := rt.modules.RootDir()

	lg := rt.Logger()
	home := sm.GetDefault(vars.BeeHome, ".bee")
	goos := sm.GetDefault(vars.BeePlatformVars, "linux")

	localDir := filepath.Join(root, bm.Root)
	remoteDir := path.Join(home, "modules", bm.Root)
	if goos == "windows" {
		remoteDir = strings.ReplaceAll(remoteDir, "/", "\\")
	}

	toSync := sm.Exists(syncFlag)
	if !toSync {
		rs, _ := conn.Stat(ctx, remoteDir)
		if rs == nil {
			toSync = true
		}

		beePath := path.Join(remoteDir, "bee.yml")
		if goos == "windows" {
			beePath = strings.ReplaceAll(beePath, "/", "\\")
		}
		data, err := conn.ReadFile(ctx, beePath)
		if err != nil {
			toSync = true
		} else {
			var om module.Module
			_ = yaml.Unmarshal(data, &om)
			if len(om.Version) != 0 && om.Version != bm.Version {
				toSync = true
			}
		}
	}

	if toSync {
		lg.Debug("put bee module",
			zap.String("name", bm.Name),
			zap.String("local", localDir),
			zap.String("remote", remoteDir))
		err := conn.Put(ctx, localDir, remoteDir, client.PutWithDir(true))
		if err != nil {
			return err
		}
	}

	return nil
}

func (rt *Runtime) Stop() error {
	rt.pool.Release()
	if err := rt.db.Flush(); err != nil {
		return err
	}
	if err := rt.db.Close(); err != nil {
		return err
	}

	return nil
}

func (rt *Runtime) applyStableMap(host string) *module.StableMap {
	sm := module.NewVariables()
	home := rt.variables.MustGetHostDefaultValue(host, vars.BeeHome, ".bee")
	sm.Set(vars.BeeHome, home)
	goos := rt.variables.MustGetHostDefaultValue(host, vars.BeePlatformVars, "linux")
	sm.Set(vars.BeePlatformVars, goos)
	arch := rt.variables.MustGetHostDefaultValue(host, vars.BeeArchVars, "amd64")
	sm.Set(vars.BeeArchVars, arch)
	return sm
}
