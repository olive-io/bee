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

package grpc

import (
	"bytes"
	"context"
	"io"
	"os"
	"sync"

	"github.com/cockroachdb/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/olive-io/bee/api/rpc"
	"github.com/olive-io/bee/api/rpctype"
	"github.com/olive-io/bee/executor/client"
)

var (
	ErrAlreadyStarted = errors.New("cmd already started")
	ErrNotStarted     = errors.New("cmd not started")
)

type cmdReader struct {
	s pb.RemoteRPC_ExecuteClient
}

func (r *cmdReader) Write(data []byte) (n int, err error) {
	n = len(data)
	err = r.s.Send(&pb.ExecuteRequest{Data: data})
	return
}

func (r *cmdReader) Close() error {
	return r.s.CloseSend()
}

type Cmd struct {
	lg *zap.Logger

	ctx     context.Context
	cancel  context.CancelFunc
	options []grpc.CallOption

	cc pb.RemoteRPCClient
	s  pb.RemoteRPC_ExecuteClient

	name string
	args []string
	envs map[string]string

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	wgMu sync.RWMutex
	wg   sync.WaitGroup

	stopping chan struct{}
	done     chan struct{}
	stop     chan struct{}
}

func (c *Cmd) goroutine(fn func()) {
	c.wgMu.RLock() // this blocks with ongoing close(s.stopping)
	defer c.wgMu.RUnlock()
	select {
	case <-c.stopping:
		c.lg.Warn("server has stopped; skipping GoAttach")
		return
	default:
	}

	// now safe to add since waitgroup wait has not started yet
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		fn()
	}()
}

func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	if c.stdin != nil {
		return nil, errors.New("exec: Stdin already set")
	}
	if c.s != nil {
		return nil, ErrAlreadyStarted
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	c.stdin = pr
	return pw, nil
}

func (c *Cmd) StdoutPipe() (io.Reader, error) {
	if c.stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.s != nil {
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
	if c.s != nil {
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
	if c.s != nil {
		return ErrAlreadyStarted
	}

	var err error
	c.s, err = c.cc.Execute(c.ctx, c.options...)
	if err != nil {
		return rpctype.ParseGRPCErr(err)
	}

	req := &pb.ExecuteRequest{
		Name: c.name,
		Args: c.args,
		Envs: c.envs,
	}
	if err = c.s.Send(req); err != nil {
		return rpctype.ParseGRPCErr(err)
	}

	rsp, err := c.s.Recv()
	if err != nil {
		return rpctype.ToGRPCErr(err)
	}
	if len(rsp.Stderr) != 0 {
		return errors.Wrap(client.ErrRequest, string(rsp.Stderr))
	}

	var b singleWriter
	if c.stdout == nil {
		c.stdout = &b
	}
	if c.stderr == nil {
		c.stderr = &b
	}

	cw := &cmdReader{s: c.s}

	c.goroutine(func() {
		defer cw.Close()
		if c.stdin == nil {
			return
		}

		_, _ = io.Copy(cw, c.stdin)
	})

	c.goroutine(func() {
		for {
			select {
			case <-c.stopping:
				return
			default:
			}

			rsp, e1 := c.s.Recv()
			if rsp != nil {
				if rsp.Stdout != nil {
					c.stdout.Write(rsp.Stdout)
				}
				if rsp.Stderr != nil {
					c.stderr.Write(rsp.Stderr)
				}
			}

			if e1 != nil {
				if e1 != io.EOF {
					c.stderr.Write([]byte(e1.Error()))
				}
				break
			}
		}

		close(c.stop)
	})

	return nil
}

func (c *Cmd) Wait() error {
	if c.s == nil {
		return ErrNotStarted
	}

	defer func() {
		c.wgMu.Lock() // block concurrent waitgroup adds in GoAttach while stopping
		close(c.stopping)
		c.wgMu.Unlock()

		c.wg.Wait()

		close(c.done)
	}()

	<-c.stop
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
	w.mu.Lock()
	defer w.mu.Unlock()
	w.b.Reset()
	return nil
}

func (c *Cmd) CombinedOutput() ([]byte, error) {
	if c.s != nil {
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
	if c.s == nil {
		return ErrNotStarted
	}

	c.cancel()
	<-c.stop
	return nil
}
