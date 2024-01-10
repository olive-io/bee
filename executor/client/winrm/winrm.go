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

package winrm

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/masterzen/winrm"
	"go.uber.org/zap"

	"github.com/olive-io/bee/executor/client"
)

type WinRM struct {
	cfg Config

	cc *winrm.Client
}

func NewWinRM(cfg Config) (*WinRM, error) {
	var err error
	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	lg := cfg.Logger
	wr := &WinRM{
		cfg: cfg,
	}
	lg.Debug("connect to remote windows",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("user", cfg.Username))
	wr.cc, err = wr.dial()
	if err != nil {
		return nil, err
	}

	return wr, nil
}

func (wr *WinRM) dial() (*winrm.Client, error) {
	cfg := wr.cfg
	cc, err := winrm.NewClient(&cfg.Endpoint, cfg.Username, cfg.Password)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

func (wr *WinRM) Name() string {
	return client.WinRMClient
}

func (wr *WinRM) Get(ctx context.Context, src, dst string, opts ...client.GetOption) error {
	options := client.NewGetOptions()
	for _, opt := range opts {
		opt(options)
	}

	info, err := fetchRemoteDir(ctx, wr.cc, src)
	if err != nil {
		return err
	}
	if len(info) == 0 {
		return nil
	}

	buf := make([]byte, options.CacheSize)
	traceFn := options.Trace
	if len(info) == 1 {
		return wr.get(ctx, src, dst, buf, traceFn)
	}

	_ = os.MkdirAll(dst, 0755)
	return wr.walker(ctx, info, src, dst, buf, traceFn)
}

func (wr *WinRM) get(ctx context.Context, remotePath, localPath string, buf []byte, fn client.IOTraceFn) error {
	writer, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer writer.Close()

	return readContent(ctx, wr.cfg.Logger, wr.cc, remotePath, writer, buf, fn)
}

func (wr *WinRM) walker(ctx context.Context, items []os.FileInfo, root, local string, buf []byte, fn client.IOTraceFn) error {
	for _, item := range items {
		pth := root + "\\" + item.Name()
		if item.IsDir() {
			dirs, _ := fetchRemoteDir(ctx, wr.cc, pth)
			if len(dirs) != 0 {
				lpth := filepath.Join(local, item.Name())
				_ = os.MkdirAll(lpth, 0755)
				if err := wr.walker(ctx, dirs, pth, lpth, buf, fn); err != nil {
					return err
				}
			}
			continue
		}
		dst := filepath.Join(local, item.Name())
		if err := wr.get(ctx, pth, dst, buf, fn); err != nil {
			return err
		}
	}
	return nil
}

func (wr *WinRM) Put(ctx context.Context, src, dst string, opts ...client.PutOption) error {
	options := client.NewPutOptions()
	for _, opt := range opts {
		opt(options)
	}

	return wr.Copy(ctx, src, dst, options.Trace)
}

func (wr *WinRM) Execute(ctx context.Context, shell string, opts ...client.ExecOption) (client.ICmd, error) {
	options := client.NewExecOptions()
	for _, opt := range opts {
		opt(options)
	}

	bash, err := wr.cc.CreateShell()
	if err != nil {
		return nil, errors.Wrap(client.ErrRequest, err.Error())
	}

	cmd := &Cmd{
		ctx:           ctx,
		name:          shell,
		args:          options.Args,
		envs:          options.Environments,
		s:             bash,
		childIOFiles:  make([]io.Closer, 0),
		parentIOPipes: make([]io.Closer, 0),
	}
	return cmd, nil
}

func (wr *WinRM) Copy(ctx context.Context, fromPath, toPath string, fn client.IOTraceFn) error {
	f, err := os.Open(fromPath)
	if err != nil {
		return errors.Wrapf(err, "couldn't read file %s", fromPath)
	}

	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return errors.Wrapf(err, "couldn't stat file %s", fromPath)
	}

	if !fi.IsDir() {
		return wr.Write(ctx, toPath, f, fn)
	} else {
		fw := fileWalker{
			ctx:     ctx,
			cc:      wr.cc,
			toDir:   toPath,
			fromDir: fromPath,
			fn:      fn,
		}
		return filepath.Walk(fromPath, fw.copyFile)
	}
}

func (wr *WinRM) Write(ctx context.Context, toPath string, src *os.File, fn client.IOTraceFn) error {
	return doCopy(ctx, wr.cfg.Logger, wr.cc, src, winPath(toPath), fn)
}

func (wr *WinRM) Close() error {
	return nil
}
