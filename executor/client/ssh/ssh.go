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

package ssh

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/olive-io/bee/executor/client"
)

type Client struct {
	cfg Config

	sc *ssh.Client
}

func NewClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	c := &Client{
		cfg: cfg,
	}

	sc, err := c.dial()
	if err != nil {
		return nil, err
	}
	c.sc = sc

	return c, nil
}

func (c *Client) dial() (*ssh.Client, error) {
	network := c.cfg.Network
	addr := c.cfg.Addr
	ccfg := c.cfg.ClientConfig

	sc, err := ssh.Dial(network, addr, ccfg)
	if err != nil {
		return nil, errors.Wrap(client.ErrConnect, err.Error())
	}
	return sc, nil
}

func (c *Client) newSFTPSession() (*sftp.Client, error) {
	copts := []sftp.ClientOption{}
	sfc, err := sftp.NewClient(c.sc, copts...)
	if err != nil {
		return nil, err
	}
	return sfc, nil
}

func (c *Client) Name() string {
	return client.SSHClient
}

func (c *Client) Stat(ctx context.Context, name string) (*client.Stat, error) {
	sfc, err := c.newSFTPSession()
	if err != nil {
		return nil, errors.Wrap(client.ErrRequest, err.Error())
	}

	lstat, err := sfc.Stat(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.Wrapf(client.ErrNotExists, err.Error())
	}
	stat := &client.Stat{
		Name:    lstat.Name(),
		IsDir:   lstat.IsDir(),
		Mod:     lstat.Mode(),
		ModTime: lstat.ModTime(),
		Size:    lstat.Size(),
	}

	return stat, nil
}

func (c *Client) ReadFile(ctx context.Context, name string) ([]byte, error) {
	sfc, err := c.newSFTPSession()
	if err != nil {
		return nil, errors.Wrap(client.ErrRequest, err.Error())
	}

	f, err := sfc.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return data, nil
}

func (c *Client) Get(ctx context.Context, src, dst string, opts ...client.GetOption) error {
	options := client.NewGetOptions()
	for _, opt := range opts {
		opt(options)
	}

	sfc, err := c.newSFTPSession()
	if err != nil {
		return errors.Wrap(client.ErrRequest, err.Error())
	}

	stat, err := sfc.Stat(src)
	if errors.Is(err, os.ErrNotExist) {
		return errors.Wrapf(client.ErrNotExists, err.Error())
	}

	buf := make([]byte, options.CacheSize)

	if stat.IsDir() {
		lstat, e1 := os.Stat(dst)
		if e1 != nil {
			return e1
		}
		if !lstat.IsDir() {
			return errors.Wrap(client.ErrAlreadyExists, dst)
		}

		walker := sfc.Walk(src)
		for walker.Step() {
			if walker.Err() != nil || walker.Path() == src {
				continue
			}

			sub := strings.TrimPrefix(walker.Path(), src)
			if walker.Stat().IsDir() {
				_ = os.MkdirAll(filepath.Join(dst, sub), os.ModePerm)
				continue
			}
			dest := filepath.Join(dst, sub)
			if _, err = get(ctx, sfc, walker.Path(), dest, buf, options.Trace); err != nil {
				break
			}
		}

	} else {
		lstat, _ := os.Stat(dst)
		if lstat != nil && lstat.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
		}

		_, err = get(ctx, sfc, src, dst, buf, options.Trace)
	}
	if err != nil {
		return errors.Wrap(client.ErrRequest, err.Error())
	}

	return nil
}

func (c *Client) Put(ctx context.Context, src, dst string, opts ...client.PutOption) error {
	options := client.NewPutOptions()
	for _, opt := range opts {
		opt(options)
	}

	sfc, err := c.newSFTPSession()
	if err != nil {
		return errors.Wrap(client.ErrRequest, err.Error())
	}

	stat, err := os.Stat(src)
	if errors.Is(err, os.ErrNotExist) {
		return errors.Wrapf(client.ErrNotExists, err.Error())
	}

	buf := make([]byte, options.CacheSize)

	var ee error
	if stat.IsDir() {
		rstat, _ := sfc.Stat(dst)
		if rstat != nil && !rstat.IsDir() {
			return errors.Wrap(client.ErrAlreadyExists, dst)
		}

		ee = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == dst {
				return nil
			}

			sub := strings.TrimPrefix(path, src)
			if d.IsDir() {
				return sfc.MkdirAll(filepath.Join(dst, sub))
			}
			dest := filepath.Join(dst, sub)
			_, e1 := put(ctx, sfc, path, dest, buf, options.Trace)
			return e1
		})

	} else {
		if options.Mkdir {
			_ = sfc.MkdirAll(path.Dir(dst))
		} else {
			rstat, _ := sfc.Stat(dst)
			if rstat != nil && rstat.IsDir() {
				dst = filepath.Join(src, filepath.Base(src))
			}
		}

		_, ee = put(ctx, sfc, src, dst, buf, options.Trace)
	}

	if ee != nil {
		return errors.Wrap(client.ErrRequest, ee.Error())
	}

	return nil
}

func (c *Client) Execute(ctx context.Context, shell string, opts ...client.ExecOption) (client.ICmd, error) {
	options := client.NewExecOptions()
	for _, opt := range opts {
		opt(options)
	}

	session, err := c.sc.NewSession()
	if err != nil {
		return nil, errors.Wrap(client.ErrConnect, err.Error())
	}

	cmd := &Cmd{
		ctx:     ctx,
		session: session,
		root:    options.Root,
		name:    shell,
		args:    options.Args,
		envs:    options.Environments,
	}

	return cmd, nil
}

func (c *Client) Close() error {
	if err := c.sc.Close(); err != nil {
		return err
	}
	return nil
}
