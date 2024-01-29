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
	"os"
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
)

var (
	DefaultParallel = runtime.NumCPU() * 2
)

type Options struct {
	dir        string
	parallel   int
	check      bool
	modulePath []string
	logger     *zap.Logger
}

func newOptions() *Options {
	home, _ := os.UserHomeDir()
	options := Options{
		dir:      filepath.Join(home, ".bee"),
		parallel: DefaultParallel,
		logger:   zap.NewExample(),
	}
	return &options
}

type Option func(*Options)

func SetDir(dir string) Option {
	return func(opt *Options) {
		opt.dir = dir
	}
}

func SetParallel(parallel int) Option {
	return func(opt *Options) {
		opt.parallel = parallel
	}
}

func SetCheck(check bool) Option {
	return func(opt *Options) {
		opt.check = check
	}
}

func SetModulePath(modulePath []string) Option {
	return func(opt *Options) {
		opt.modulePath = modulePath
	}
}

func SetLogger(lg *zap.Logger) Option {
	return func(opt *Options) {
		opt.logger = lg
	}
}

type RunOptions struct {
	sync bool
}

func newRunOptions() *RunOptions {
	options := RunOptions{}
	return &options
}

type RunOption func(*RunOptions)

func SetRunSync(b bool) RunOption {
	return func(opt *RunOptions) {
		opt.sync = b
	}
}
