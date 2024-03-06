//go:build windows

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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	pb "github.com/olive-io/bee/api/rpc"
)

func startExec(ctx context.Context, in *pb.ExecuteRequest) *exec.Cmd {
	shell := fmt.Sprintf("%s %s", in.Name, strings.Join(in.Args, " "))
	cmd := exec.CommandContext(ctx, "cmd", "/C", shell)

	sysAttr := &syscall.SysProcAttr{
		HideWindow: true,
	}

	cmd.SysProcAttr = sysAttr
	cmd.Env = append(cmd.Env, os.Environ()...)
	for k, v := range in.Envs {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	home, err := os.UserHomeDir()
	if err == nil {
		cmd.Dir = home
	}

	return cmd
}
