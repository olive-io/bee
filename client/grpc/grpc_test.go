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
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "github.com/olive-io/bee/api/rpc"
	bs "github.com/olive-io/bee/server/grpc"
)

func newClient(t *testing.T) *Client {
	lg := zap.NewExample()
	rand.NewSource(time.Now().Unix())
	port := rand.Intn(5000) + 10000
	addr := fmt.Sprintf("localhost:%d", port)

	kp := keepalive.ServerParameters{
		Time:    5 * time.Minute,
		Timeout: 1 * time.Minute,
	}

	impl := bs.NewServer()
	server := grpc.NewServer(grpc.KeepaliveParams(kp))
	pb.RegisterRemoteRPCServer(server, impl)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_ = server.Serve(ln)
	}()

	time.Sleep(time.Second)
	cfg := NewConfig(lg, addr)
	client, err := NewClient(*cfg)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestNewClient(t *testing.T) {
	client := newClient(t)
	err := client.Close()
	if !assert.NoError(t, err) {
		return
	}
}

func TestClient_Put_Get(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	ctx := context.Background()
	tf := filepath.Join(os.TempDir(), "test.txt")
	err := os.WriteFile(tf, []byte("hello world1"), os.ModePerm)
	if !assert.NoError(t, err) {
		return
	}

	err = c.Put(ctx, tf, "/tmp/test.txt")
	if !assert.NoError(t, err) {
		return
	}

	local := filepath.Join(os.TempDir(), "test1.txt")
	err = c.Get(ctx, "/tmp/test.txt", local)
	if !assert.NoError(t, err) {
		return
	}

	data, _ := os.ReadFile(local)
	if !assert.Equal(t, []byte("hello world1"), data) {
		return
	}
}

func TestClient_Execute(t *testing.T) {
	c := newClient(t)
	defer c.Close()

	ctx := context.Background()
	cmd, err := c.Execute(ctx, "ifconfig")
	if !assert.NoError(t, err) {
		return
	}
	defer cmd.Close()

	data, err := cmd.CombinedOutput()
	if !assert.NoError(t, err) {
		return
	}

	t.Logf("%v", string(data))
}
