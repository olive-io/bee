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
	"context"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pb "github.com/olive-io/bee/api/rpc"
	"github.com/olive-io/bee/api/rpctype"
	"github.com/olive-io/bee/client"
)

type Client struct {
	cfg Config
	cc  pb.RemoteRPCClient
}

func NewClient(cfg Config) (*Client, error) {
	var err error
	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	c := &Client{
		cfg: cfg,
	}

	conn := cfg.conn
	if conn != nil {
		ctx := context.Background()
		conn, err = c.dial(ctx)
		if err != nil {
			return nil, err
		}
		cfg.conn = conn
	}
	c.cc = pb.NewRemoteRPCClient(conn)

	return c, nil
}

func (c *Client) dial(ctx context.Context) (*grpc.ClientConn, error) {
	cfg := c.cfg
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	ckp := keepalive.ClientParameters{
		Time:                5 * time.Minute,
		Timeout:             1 * time.Minute,
		PermitWithoutStream: true,
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(ckp),
	}

	conn, err := grpc.DialContext(ctx, cfg.Address, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Client) callOptions() []grpc.CallOption {
	options := make([]grpc.CallOption, 0)
	return options
}

func (c *Client) stat(ctx context.Context, name string) (*pb.FileStat, error) {
	copts := c.callOptions()
	in := &pb.StatRequest{Name: name}
	rsp, err := c.cc.Stat(ctx, in, copts...)
	if err != nil {
		return nil, rpctype.ParseGRPCErr(err)
	}
	return rsp.Stat, nil
}

func (c *Client) Get(ctx context.Context, src, dst string, opts ...client.GetOption) error {
	options := client.NewGetOptions()
	for _, opt := range opts {
		opt(options)
	}

	stat, err := c.stat(ctx, src)
	if err != nil && !errors.Is(err, client.ErrNotExists) {
		return err
	}

	if stat != nil {
		dstat, _ := os.Stat(dst)
		if dstat != nil && stat.IsDir && !dstat.IsDir() {
			return errors.Wrapf(os.ErrInvalid, "not a regular file")
		}
	}

	in := &pb.GetRequest{
		Name:      src,
		CacheSize: options.CacheSize,
	}
	copts := c.callOptions()
	rc, err := c.cc.Get(ctx, in, copts...)
	if err != nil {
		return rpctype.ParseGRPCErr(err)
	}

	var fw *os.File
	fn := options.Trace
	written := int64(0)
	sub := int64(0)
	last := time.Now()
	for {
		rsp, e1 := rc.Recv()
		if e1 != nil && e1 != io.EOF {
			return rpctype.ParseGRPCErr(err)
		}

		if rs := rsp.Stat; rs != nil {
			name := rs.Name
			entry := filepath.Join(dst, strings.TrimPrefix(name, src))
			if rs.IsDir {
				if err = os.Mkdir(entry, fs.FileMode(rs.Perm)); err != nil {
					return err
				}
			} else {
				if fw == nil || fw.Name() != entry {
					if fw != nil {
						_ = fw.Close()
					}
					fw, err = os.Create(entry)
					if err != nil {
						return err
					}
				}
				chunk := rsp.Chunk
				if chunk == nil {
					return errors.Wrapf(client.ErrInvalidWrite, "no data")
				}

				trace := &client.IOTrace{
					Name:  path.Base(fw.Name()),
					Src:   name,
					Dst:   fw.Name(),
					Total: rs.Size,
				}

				var n int
				n, err = fw.Write(chunk.Data[:chunk.Length])
				if err != nil {
					return err
				}
				written += int64(n)
				if fn != nil {
					now := time.Now()
					trace.Chunk = written
					trace.Speed = int64(float64(written-sub) / (now.Sub(last).Seconds()))
					last = now
					sub = written
					fn(trace)
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

func (c *Client) Put(ctx context.Context, src, dst string, opts ...client.PutOption) error {
	options := client.NewPutOptions()
	for _, opt := range opts {
		opt(options)
	}

	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	lstat, err := c.stat(ctx, dst)
	if err != nil && !errors.Is(err, client.ErrNotExists) {
		return err
	}

	if lstat != nil && lstat.IsDir && !stat.IsDir() {
		return errors.Wrapf(os.ErrInvalid, "not a regular file")
	}

	stream, err := c.cc.Put(ctx, c.callOptions()...)
	if err != nil {
		return rpctype.ParseGRPCErr(err)
	}

	buf := make([]byte, options.CacheSize)
	fn := options.Trace
	if !stat.IsDir() {
		return c.put(ctx, stream, src, dst, buf, fn)
	}

	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dst {
			return nil
		}

		sub := strings.TrimPrefix(path, src)
		dest := filepath.Join(dst, sub)
		return c.put(ctx, stream, path, dest, buf, fn)
	})

	if err != nil {
		return err
	}

	return rpctype.ParseGRPCErr(stream.CloseSend())
}

func (c *Client) put(ctx context.Context, stream pb.RemoteRPC_PutClient, src, dst string, buf []byte, fn client.IOTraceFn) error {
	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	req := &pb.PutRequest{
		Name: dst,
		Stat: &pb.FileStat{
			Name:  stat.Name(),
			IsDir: stat.IsDir(),
			Perm:  uint32(stat.Mode()),
			Size:  stat.Size(),
		},
		Chunk: &pb.Chunk{},
	}
	if stat.IsDir() {
		return stream.Send(req)
	}

	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	var trace *client.IOTrace
	if fn != nil {
		trace = &client.IOTrace{
			Name: reader.Name(),
			Src:  src,
			Dst:  dst,
		}
		trace.Total = stat.Size()
	}

	if buf == nil {
		buf = make([]byte, 32*1024)
	}

	last := time.Now()
	sub := int64(0)
	written := int64(0)
	for {
		select {
		case <-ctx.Done():
			return client.ErrTimeout
		default:
		}

		nr, er := reader.Read(buf)
		if nr > 0 {
			written += int64(nr)
			req.Chunk = &pb.Chunk{
				Data:   buf[:nr],
				Length: int64(nr),
			}
			er = stream.Send(req)
		}
		if fn != nil {
			now := time.Now()
			trace.Chunk = written
			trace.Speed = int64(float64(written-sub) / (now.Sub(last).Seconds()))
			last = now
			sub = written
			fn(trace)
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	return err
}

func (c *Client) Execute(ctx context.Context, shell string, opts ...client.ExecOption) (client.ICmd, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Close() error {
	return c.cfg.conn.Close()
}
