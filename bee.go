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

	bexecutor "github.com/olive-io/bee/executor"
	"github.com/olive-io/bee/executor/client"
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/module"
	mmg "github.com/olive-io/bee/module/manager"
	"github.com/olive-io/bee/parser"
	"github.com/olive-io/bee/plugins/callback"
	"github.com/olive-io/bee/secret"
	"github.com/olive-io/bee/vars"
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
	callback  callback.ICallBack
}

func NewRuntime(
	inventory *inv.Manager, variables *vars.VariableManager,
	loader *parser.DataLoader, callback callback.ICallBack, opts ...Option,
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
		callback:  callback,
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

//func (rt *Runtime) Run(ctx context.Context, task *playbook.Task, opts ...RunOption) error {
//	options := newRunOptions()
//	for _, opt := range opts {
//		opt(options)
//	}
//
//	lg := rt.Logger()
//
//	mname := task.Module
//	var args []string
//	if before, after, ok := strings.Cut(mname, "."); ok {
//		mname = before
//		args = append(args, after)
//	}
//	bm, ok := rt.modules.Find(mname)
//	if !ok {
//		return errors.New("unknown module")
//	}
//
//	for key, arg := range task.Args {
//		args = append(args, key+"="+arg)
//	}
//
//	cmd, err := bm.Execute(args...)
//	if err != nil {
//		return err
//	}
//
//	if !cmd.Runnable() {
//		return errors.New("command can't be execute")
//	}
//
//	for _, host := range task.Hosts {
//		goos := rt.variables.MustGetHostDefaultValue(host, vars.BeePlatformVars, "linux")
//		home := rt.variables.MustGetHostDefaultValue(host, vars.BeeHome, ".bee")
//		conn, err := rt.executor.GetClient(host)
//		if err != nil {
//			return err
//		}
//
//		if err = rt.syncRepl(ctx, host, conn); err != nil {
//			return err
//		}
//
//		putOptions := client.NewPutOptions()
//		putOptions.Dir = true
//		mctx := context.WithValue(ctx, HomeKey, home)
//		mctx = context.WithValue(mctx, OsKey, goos)
//		putOptions.Context = mctx
//		if err = rt.syncModules(ctx, conn, putOptions); err != nil {
//			return err
//		}
//
//		execOptions := make([]client.ExecOption, 0)
//		execOptions = append(execOptions, client.ExecWithValue(HomeKey, home))
//		execOptions = append(execOptions, client.ExecWithValue(OsKey, goos))
//		data, err := rt.runCommand(ctx, cmd, conn, execOptions...)
//		if err != nil {
//			return err
//		}
//
//		lg.Info("output: " + string(data))
//	}
//
//	return nil
//}

func (rt *Runtime) Execute(ctx context.Context, host, shell string, opts ...RunOption) ([]byte, error) {
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
	if err = rt.syncRepl(ctx, conn, sm); err != nil {
		return nil, err
	}

	if err = rt.syncModules(ctx, conn, sm); err != nil {
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

func (rt *Runtime) syncRepl(ctx context.Context, conn client.IClient, sm *module.StableMap[string]) error {
	lg := rt.opts.logger
	home := sm.GetDefault(vars.BeeHome, ".bee")
	goos := sm.GetDefault(vars.BeePlatformVars, "linux")
	arch := sm.GetDefault(vars.BeeArchVars, "amd64")

	repl := path.Join(home, "bin", "tengo")
	if goos == "windows" {
		repl = strings.ReplaceAll(repl, "/", "\\")
		repl += ".exe"
	}

	cmd, err := conn.Execute(ctx, repl, client.ExecWithArgs("-version"))
	if err != nil {
		return err
	}
	_, err = cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	toolchain := filepath.Join(rt.opts.dir, "repl", "tengo."+goos+"."+arch)
	if goos == "windows" {
		toolchain += ".exe"
	}
	_, err = os.Stat(toolchain)
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

func (rt *Runtime) syncModules(ctx context.Context, conn client.IClient, sm *module.StableMap[string]) error {
	root := rt.modules.RootDir()
	dirs := rt.modules.ModuleDirs()

	lg := rt.Logger()
	home := sm.GetDefault(vars.BeeHome, ".bee")
	goos := sm.GetDefault(vars.BeePlatformVars, "linux")

	for _, dir := range dirs {
		localDir := dir
		remoteDir := path.Join(home, "modules", strings.TrimPrefix(dir, root))
		if goos == "windows" {
			remoteDir = strings.ReplaceAll(remoteDir, "/", "\\")
		}
		rs, _ := conn.Stat(ctx, remoteDir)
		if rs != nil {
			continue
		}
		lg.Debug("put bee module", zap.String("local", localDir), zap.String("remote", remoteDir))
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

func (rt *Runtime) applyStableMap(host string) *module.StableMap[string] {
	sm := module.NewVariables()
	home := rt.variables.MustGetHostDefaultValue(host, vars.BeeHome, ".bee")
	sm.Set(vars.BeeHome, home)
	goos := rt.variables.MustGetHostDefaultValue(host, vars.BeePlatformVars, "linux")
	sm.Set(vars.BeePlatformVars, goos)
	arch := rt.variables.MustGetHostDefaultValue(host, vars.BeeArchVars, "amd64")
	sm.Set(vars.BeeArchVars, arch)
	return sm
}