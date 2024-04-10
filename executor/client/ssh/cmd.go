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
	"bytes"
	"context"
	"io"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"golang.org/x/crypto/ssh"
)

type Cmd struct {
	ctx     context.Context
	session *ssh.Session

	root string
	name string
	args []string
	envs map[string]string
}

func (c *Cmd) shell() string {
	args := make([]string, 0)
	if c.root != "" {
		args = append(args, "cd "+c.root+";")
	}
	args = append(args, c.name)
	args = append(args, c.args...)
	return strings.Join(args, " ")
}

func (c *Cmd) Session() *ssh.Session {
	return c.session
}

func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	return c.session.StdinPipe()
}

func (c *Cmd) StdoutPipe() (io.Reader, error) {
	return c.session.StdoutPipe()
}

func (c *Cmd) StderrPipe() (io.Reader, error) {
	return c.session.StderrPipe()
}

func (c *Cmd) Start() error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	shell := c.shell()
	for key, value := range c.envs {
		if err := c.session.Setenv(key, value); err != nil {
			return err
		}
	}
	return c.session.Start(shell)
}

func (c *Cmd) Wait() error {
	ech := make(chan error, 1)
	go func() {
		ech <- c.session.Wait()
	}()

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case err := <-ech:
		return err
	}
}

func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

type singleWriter struct {
	b  bytes.Buffer
	mu sync.Mutex
}

func (w *singleWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.Write(p)
}

func (c *Cmd) CombinedOutput() ([]byte, error) {
	if c.session.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.session.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b singleWriter
	c.session.Stdout = &b
	c.session.Stderr = &b
	err := c.Run()
	return b.b.Bytes(), err
}

func (c *Cmd) Close() error {
	return c.session.Close()
}
