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
	"path/filepath"
	"runtime"

	"github.com/olive-io/bee/plugins/callback"
	"github.com/olive-io/bee/plugins/filter"
	"github.com/olive-io/bpmn/tracing"
	"go.uber.org/zap"
)

var (
	DefaultParallel = runtime.NumCPU() * 2
)

type Callable func(ctx context.Context, host, action string, in []byte, opts ...RunOption) ([]byte, error)

type Options struct {
	dir      string
	parallel int
	check    bool
	logger   *zap.Logger
	caller   Callable
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

func SetLogger(lg *zap.Logger) Option {
	return func(opt *Options) {
		opt.logger = lg
	}
}

func SetCaller(caller Callable) Option {
	return func(opt *Options) {
		opt.caller = caller
	}
}

type RunOptions struct {
	Callback  callback.ICallBack
	Filter    filter.IFilter
	Tracer    chan tracing.ITrace
	Metadata  map[string]any
	ExtraArgs map[string]string
	sync      bool
}

func newRunOptions() *RunOptions {
	options := RunOptions{
		Callback: callback.NewCallBack(),
		Filter:   filter.NewFilter(),
	}
	return &options
}

type RunOption func(*RunOptions)

func WithRunSync(b bool) RunOption {
	return func(opt *RunOptions) {
		opt.sync = b
	}
}

func WithRunCallback(cb callback.ICallBack) RunOption {
	return func(opt *RunOptions) {
		opt.Callback = cb
	}
}

func WithRunFilter(f filter.IFilter) RunOption {
	return func(opt *RunOptions) {
		opt.Filter = f
	}
}

func WithRunTracer(tracer chan tracing.ITrace) RunOption {
	return func(opt *RunOptions) {
		opt.Tracer = tracer
	}
}

func WithMetadata(md map[string]any) RunOption {
	return func(opt *RunOptions) {
		if opt.Metadata == nil {
			opt.Metadata = map[string]any{}
		}
		for key, value := range md {
			opt.Metadata[key] = value
		}
	}
}

func WithArgs(args map[string]string) RunOption {
	return func(opt *RunOptions) {
		if opt.ExtraArgs == nil {
			opt.ExtraArgs = map[string]string{}
		}
		for key, value := range args {
			opt.ExtraArgs[key] = value
		}
	}
}
