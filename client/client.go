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

package client

import (
	"context"
	"io"
	"time"
)

const (
	DefaultCacheSize   = 1024 * 32
	DefaultDialTimeout = time.Second * 15
)

type IClient interface {
	// Get gets io.ReadCloser from remote connection
	Get(ctx context.Context, src, dst string, opts ...GetOption) error
	// Put uploads local file to remote host
	Put(ctx context.Context, src, dst string, opts ...PutOption) error
	// Execute executes a given command on remote host
	Execute(ctx context.Context, shell string, opts ...ExecOption) (ICmd, error)
	// Close closes all remote connections
	Close() error
}

type ICmd interface {
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	Start() error
	Wait() error
	Run() error
	CombinedOutput() ([]byte, error)
	Close() error
}

type GetOptions struct {
	Dir       bool
	CacheSize int64
	Trace     IOTraceFn
}

func NewGetOptions() *GetOptions {
	opt := &GetOptions{
		CacheSize: DefaultCacheSize,
	}
	return opt
}

type GetOption func(*GetOptions)

func GetWithDir(dir bool) GetOption {
	return func(options *GetOptions) {
		options.Dir = dir
	}
}

func GetWithTrace(trace IOTraceFn) GetOption {
	return func(options *GetOptions) {
		options.Trace = trace
	}
}

type PutOptions struct {
	Dir       bool
	CacheSize int64
	Trace     IOTraceFn
}

func NewPutOptions() *PutOptions {
	opt := &PutOptions{
		CacheSize: DefaultCacheSize,
	}
	return opt
}

type PutOption func(*PutOptions)

func PutWithDir(dir bool) PutOption {
	return func(options *PutOptions) {
		options.Dir = dir
	}
}

func PutWithTrace(trace IOTraceFn) PutOption {
	return func(options *PutOptions) {
		options.Trace = trace
	}
}

type IOTrace struct {
	Name  string
	Src   string
	Dst   string
	Total int64
	Chunk int64
	Speed int64
}

type IOTraceFn func(*IOTrace)

type ExecOptions struct {
	Args         []string
	Environments map[string]string
}

func NewExecOptions() *ExecOptions {
	opt := &ExecOptions{
		Environments: map[string]string{},
	}
	return opt
}

type ExecOption func(*ExecOptions)

func ExecWithArgs(args []string) ExecOption {
	return func(options *ExecOptions) {
		options.Args = args
	}
}

func ExecWithEnv(key, value string) ExecOption {
	return func(options *ExecOptions) {
		if options.Environments == nil {
			options.Environments = map[string]string{}
		}
		options.Environments[key] = value
	}
}
