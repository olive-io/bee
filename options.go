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

package bee

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/olive-io/bpmn/tracing"
	"go.uber.org/zap"

	"github.com/olive-io/bee/plugins/callback"
	"github.com/olive-io/bee/plugins/filter"
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
