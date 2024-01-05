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
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	pb "github.com/olive-io/bee/api/rpc"
	"github.com/olive-io/bee/api/rpctype"
)

type Server struct{}

func NewServer() *Server {
	s := &Server{}
	return s
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
	defer func() {
		if fw != nil {
			_ = fw.Close()
		}
	}()

	for {
		req, e1 := stream.Recv()
		if err := s.save(req, fw); err != nil {
			return rpctype.ToGRPCErr(err)
		}
		if e1 != nil {
			if e1 == io.EOF {
				break
			}
			return rpctype.ToGRPCErr(e1)
		}
	}

	return nil
}

func (s *Server) save(req *pb.PutRequest, fw *os.File) error {
	if req == nil || req.Stat == nil {
		return nil
	}

	stat := req.Stat
	if stat.IsDir {
		mod := fs.FileMode(stat.Perm)
		if mod == 0 {
			mod = os.ModePerm
		}
		return os.Mkdir(req.Name, mod)
	}

	var err error
	if fw == nil || fw.Name() != req.Name {
		if fw != nil {
			_ = fw.Close()
		}
		fw, err = os.Create(req.Name)
		if err != nil {
			return err
		}
	}

	if chunk := req.Chunk; chunk != nil {
		data := chunk.Data[:chunk.Length]
		_, err = fw.Write(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Execute(stream pb.RemoteRPC_ExecuteServer) error {
	ctx := stream.Context()
	req, err := stream.Recv()
	if err != nil {
		return rpctype.ToGRPCErr(err)
	}

	name := req.Name
	args := req.Args
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = make([]string, 0)
	for key, value := range req.Envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

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

	if err = cmd.Wait(); err != nil {
		return rpctype.ToGRPCErr(cmd.Wait())
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
	rsp := &pb.ExecuteResponse{Stdout: data}
	if err := s.s.Send(rsp); err != nil {
		return 0, rpctype.ToGRPCErr(err)
	}
	return len(data), nil
}

type execStderr struct {
	s pb.RemoteRPC_ExecuteServer
}

func (s *execStderr) Write(data []byte) (int, error) {
	rsp := &pb.ExecuteResponse{Stderr: data}
	if err := s.s.Send(rsp); err != nil {
		return 0, rpctype.ToGRPCErr(err)
	}
	return len(data), nil
}
