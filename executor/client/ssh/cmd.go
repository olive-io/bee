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
