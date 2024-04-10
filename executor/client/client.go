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

package client

import (
	"context"
	"io"
	"os"
	"time"
)

const (
	DefaultCacheSize   = 1024 * 1024
	DefaultDialTimeout = time.Minute * 3
	DefaultExecTimeout = time.Minute * 10
)

const (
	SSHClient   = "ssh"
	WinRMClient = "winrm"
	GRPCClient  = "grpc"
)

type IClient interface {
	Name() string
	Stat(ctx context.Context, name string) (*Stat, error)
	// ReadFile reads file content from remote connection
	ReadFile(ctx context.Context, name string) ([]byte, error)
	// Get gets io.ReadCloser from remote connection
	Get(ctx context.Context, src, dst string, opts ...GetOption) error
	// Put uploads local file to remote host
	Put(ctx context.Context, src, dst string, opts ...PutOption) error
	// Execute executes a given command on remote host
	Execute(ctx context.Context, shell string, opts ...ExecOption) (ICmd, error)
	// Close closes all remote connections
	Close() error
}

type Stat struct {
	Name    string
	IsDir   bool
	Mod     os.FileMode
	ModTime time.Time
	Size    int64
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
	Context context.Context

	Dir       bool
	CacheSize int64
	Trace     IOTraceFn
}

func NewGetOptions() *GetOptions {
	opt := &GetOptions{
		Context:   context.TODO(),
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

func GetWithValue(key string, value any) GetOption {
	return func(options *GetOptions) {
		options.Context = context.WithValue(options.Context, key, value)
	}
}

type PutOptions struct {
	Context context.Context

	Dir       bool
	Mkdir     bool
	CacheSize int64
	Trace     IOTraceFn
}

func NewPutOptions() *PutOptions {
	opt := &PutOptions{
		Context:   context.TODO(),
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

func PutWithMkdir(mkdir bool) PutOption {
	return func(options *PutOptions) {
		options.Mkdir = mkdir
	}
}

func PutWithTrace(trace IOTraceFn) PutOption {
	return func(options *PutOptions) {
		options.Trace = trace
	}
}

func PutWithValue(key string, value any) PutOption {
	return func(options *PutOptions) {
		options.Context = context.WithValue(options.Context, key, value)
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
	Context context.Context

	Root         string
	Args         []string
	Environments map[string]string
	Timeout      time.Duration
}

func NewExecOptions() *ExecOptions {
	opt := &ExecOptions{
		Context:      context.TODO(),
		Environments: map[string]string{},
		Timeout:      DefaultExecTimeout,
	}
	return opt
}

type ExecOption func(*ExecOptions)

func ExecWithRootDir(root string) ExecOption {
	return func(options *ExecOptions) {
		options.Root = root
	}
}

func ExecWithArgs(args ...string) ExecOption {
	return func(options *ExecOptions) {
		options.Args = append(options.Args, args...)
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

func ExecWithTimeout(timeout time.Duration) ExecOption {
	return func(options *ExecOptions) {
		options.Timeout = timeout
	}
}

// ExecWithValue set key-value at options.Context
func ExecWithValue(key string, value any) ExecOption {
	return func(options *ExecOptions) {
		options.Context = context.WithValue(options.Context, key, value)
	}
}
