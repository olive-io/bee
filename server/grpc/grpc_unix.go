//go:build !windows

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

package grpc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	pb "github.com/olive-io/bee/api/rpc"
)

func startExec(ctx context.Context, in *pb.ExecuteRequest) *exec.Cmd {
	shell := fmt.Sprintf("%s %s", in.Name, strings.Join(in.Args, " "))

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", shell)
	sysAttr := &syscall.SysProcAttr{
		Setpgid: true,
	}

	sysAttr.Credential = &syscall.Credential{
		Uid: uint32(0),
		Gid: uint32(0),
	}

	cmd.SysProcAttr = sysAttr
	cmd.Env = append(cmd.Env, os.Environ()...)
	for k, v := range in.Envs {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if in.Root != "" {
		cmd.Dir = in.Root
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			cmd.Dir = home
		}
	}

	return cmd
}
