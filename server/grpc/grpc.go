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
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"

	pb "github.com/olive-io/bee/api/rpc"
	"github.com/olive-io/bee/api/rpctype"
)

type Server struct{}

func NewServer() *Server {
	s := &Server{}
	return s
}

func RegisterGRPCBeeHandler(s *grpc.Server) {
	pb.RegisterRemoteRPCServer(s, NewServer())
}

func (s *Server) Stat(ctx context.Context, req *pb.StatRequest) (*pb.StatResponse, error) {
	stat, err := os.Stat(req.Name)
	if err != nil {
		return nil, rpctype.ToGRPCErr(err)
	}

	resp := &pb.StatResponse{
		Stat: &pb.FileStat{
			Name:    stat.Name(),
			IsDir:   stat.IsDir(),
			Perm:    uint32(stat.Mode()),
			Size:    stat.Size(),
			ModTime: stat.ModTime().Unix(),
		},
	}
	return resp, nil
}

func (s *Server) Get(req *pb.GetRequest, stream pb.RemoteRPC_GetServer) error {
	stat, err := os.Stat(req.Name)
	if err != nil {
		return rpctype.ToGRPCErr(err)
	}

	size := req.CacheSize
	if size == 0 {
		size = 1024 * 32
	}
	buf := make([]byte, size)
	if !stat.IsDir() {
		return s.get(stream, req.Name, buf)
	}

	err = filepath.WalkDir(req.Name, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			lstat := &pb.FileStat{
				Name:  d.Name(),
				IsDir: d.IsDir(),
			}
			if info, _ := d.Info(); info != nil {
				lstat.ModTime = info.ModTime().Unix()
				lstat.Perm = uint32(info.Mode())
			}
			rsp := &pb.GetResponse{
				Stat:  lstat,
				Chunk: &pb.Chunk{},
			}
			return rpctype.ToGRPCErr(stream.Send(rsp))
		}
		return s.get(stream, path, buf)
	})

	return nil
}

func (s *Server) get(stream pb.RemoteRPC_GetServer, name string, buf []byte) error {
	reader, err := os.Open(name)
	if err != nil {
		return rpctype.ToGRPCErr(err)
	}
	defer reader.Close()

	rs, err := reader.Stat()
	if err != nil {
		return rpctype.ToGRPCErr(err)
	}
	stat := &pb.FileStat{
		Name:    rs.Name(),
		IsDir:   rs.IsDir(),
		Perm:    uint32(rs.Mode()),
		Size:    rs.Size(),
		ModTime: rs.ModTime().Unix(),
	}
	for {
		nr, e1 := reader.Read(buf)
		if nr > 0 {
			rsp := &pb.GetResponse{
				Stat: stat,
				Chunk: &pb.Chunk{
					Data:   buf[:nr],
					Length: int64(nr),
				},
			}
			if err = stream.Send(rsp); err != nil {
				return rpctype.ToGRPCErr(err)
			}
		}

		if e1 != nil {
			if e1 == io.EOF {
				break
			}
			return rpctype.ToGRPCErr(err)
		}
	}
	return nil
}

func (s *Server) Put(stream pb.RemoteRPC_PutServer) error {
	var fw *os.File

	for {
		req, e1 := stream.Recv()
		if e1 != nil && e1 != io.EOF {
			return rpctype.ToGRPCErr(e1)
		}

		if req != nil {
			stat := req.Stat
			if stat.IsDir {
				mod := fs.FileMode(stat.Perm)
				if mod == 0 {
					mod = os.ModePerm
				}
				if err := os.MkdirAll(req.Name, mod); err != nil {
					return rpctype.ToGRPCErr(err)
				}
			} else {
				var err error
				if fw == nil || fw.Name() != req.Name {
					if fw != nil {
						_ = fw.Close()
					}
					dir := filepath.Dir(req.Name)
					perm := os.FileMode(stat.Perm)
					if perm == 0 {
						perm = 0755
					}
					if _, err = os.Stat(dir); errors.Is(err, os.ErrNotExist) {
						_ = os.MkdirAll(dir, perm)
					}
					fw, err = os.OpenFile(req.Name, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_SYNC, perm)
					if err != nil {
						return rpctype.ToGRPCErr(err)
					}
				}

				if req != nil && req.Chunk != nil {
					chunk := req.Chunk
					data := chunk.Data[:chunk.Length]
					_, err = fw.Write(data)
					if err != nil {
						return rpctype.ToGRPCErr(err)
					}
				}
			}
		}

		if e1 == io.EOF {
			break
		}
	}

	if fw != nil {
		_ = fw.Close()
	}

	if err := stream.SendAndClose(&pb.PutResponse{}); err != nil {
		return err
	}

	return nil
}

func (s *Server) Execute(stream pb.RemoteRPC_ExecuteServer) error {
	ctx := stream.Context()
	req, err := stream.Recv()
	if err != nil {
		return rpctype.ToGRPCErr(err)
	}

	cmd := startExec(ctx, req)
	rsp := &pb.ExecuteResponse{}
	if err = stream.Send(rsp); err != nil {
		return rpctype.ToGRPCErr(err)
	}

	cmd.Stdin = &execReader{s: stream}
	cmd.Stdout = &execStdout{s: stream}
	cmd.Stderr = &execStderr{s: stream}
	if err = cmd.Start(); err != nil {
		return rpctype.ToGRPCErr(err)
	}

	go func() {
		timer := time.NewTicker(time.Second * 5)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
			}

			_ = stream.Send(&pb.ExecuteResponse{Kind: pb.ExecuteResponse_Ping})
		}
	}()

	if err = cmd.Wait(); err != nil {
		return rpctype.ToGRPCErr(err)
	}

	return nil
}

type execReader struct {
	s pb.RemoteRPC_ExecuteServer
}

func (r *execReader) Read(data []byte) (n int, err error) {
	req, err := r.s.Recv()
	if err != nil {
		return 0, rpctype.ToGRPCErr(err)
	}
	return bytes.NewBuffer(data).Write(req.Data)
}

type execStdout struct {
	s pb.RemoteRPC_ExecuteServer
}

func (s *execStdout) Write(data []byte) (int, error) {
	rsp := &pb.ExecuteResponse{Kind: pb.ExecuteResponse_Data, Stdout: data}
	if err := s.s.Send(rsp); err != nil {
		return 0, rpctype.ToGRPCErr(err)
	}
	return len(data), nil
}

type execStderr struct {
	s pb.RemoteRPC_ExecuteServer
}

func (s *execStderr) Write(data []byte) (int, error) {
	rsp := &pb.ExecuteResponse{Kind: pb.ExecuteResponse_Data, Stderr: data}
	if err := s.s.Send(rsp); err != nil {
		return 0, rpctype.ToGRPCErr(err)
	}
	return len(data), nil
}
