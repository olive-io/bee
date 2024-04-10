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

package winrm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/olive-io/winrm"

	"github.com/olive-io/bee/executor/client"
)

var (
	ErrAlreadyStarted = errors.New("cmd already started")
	ErrNotStarted     = errors.New("cmd not started")
)

type Cmd struct {
	ctx context.Context

	root string
	name string
	args []string
	envs map[string]string

	s      *winrm.Shell
	c      *winrm.Command
	stdin  io.Reader
	stdout io.WriteCloser
	stderr io.WriteCloser

	childIOFiles  []io.Closer
	parentIOPipes []io.Closer

	wg sync.WaitGroup
}

func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	if c.stdin != nil {
		return nil, errors.New("exec: Stdin already set")
	}
	if c.c != nil {
		return nil, ErrAlreadyStarted
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	c.stdin = pr
	//c.childIOFiles = append(c.childIOFiles, pr)
	//c.parentIOPipes = append(c.parentIOPipes, pw)
	return pw, nil
}

func (c *Cmd) StdoutPipe() (io.Reader, error) {
	if c.stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.c != nil {
		return nil, ErrAlreadyStarted
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	c.stdout = pw
	return pr, nil
}

func (c *Cmd) StderrPipe() (io.Reader, error) {
	if c.stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	if c.c != nil {
		return nil, ErrAlreadyStarted
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	c.stderr = pw
	return pr, nil
}

func (c *Cmd) Start() error {
	if c.c != nil {
		return ErrAlreadyStarted
	}

	ctx := c.ctx
	args := make([]string, 0)
	if c.root != "" {
		args = append(args, "Set-Location "+c.root+";")
	}
	args = append(args, c.name)
	args = append(args, c.args...)
	shell := strings.Join(args, " ")
	cc, err := c.s.ExecuteWithContext(ctx, fmt.Sprintf(`powershell -c "%s"`, shell))
	if err != nil {
		return errors.Wrapf(client.ErrRequest, err.Error())
	}

	var b singleWriter
	if c.stdout == nil {
		c.stdout = &b
	}
	if c.stderr == nil {
		c.stderr = &b
	}

	c.wg.Add(3)
	go func() {
		if c.stdin == nil {
			c.wg.Done()
			return
		}

		defer func() {
			cc.Stdin.Close()
			c.wg.Done()
		}()
		io.Copy(cc.Stdin, c.stdin)
	}()
	go func() {
		defer c.wg.Done()
		io.Copy(c.stdout, cc.Stdout)
	}()
	go func() {
		defer c.wg.Done()
		io.Copy(c.stderr, cc.Stderr)
	}()

	c.c = cc

	return nil
}

func (c *Cmd) Wait() error {
	if c.c == nil {
		return ErrNotStarted
	}
	c.c.Wait()
	c.wg.Wait()

	if err := c.c.Close(); err != nil {
		return err
	}

	if code := c.c.ExitCode(); code != 0 {
		return fmt.Errorf("code %d", code)
	}
	return nil
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

func (w *singleWriter) Close() error {
	return nil
}

func (c *Cmd) CombinedOutput() ([]byte, error) {
	if c.c != nil {
		return nil, ErrAlreadyStarted
	}

	if c.stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b singleWriter
	c.stdout = &b
	c.stderr = &b
	err := c.Run()
	return b.b.Bytes(), err
}

func (c *Cmd) Close() error {
	return c.s.Close()
}

func closeDescriptors(closers []io.Closer) {
	for _, fd := range closers {
		fd.Close()
	}
}
