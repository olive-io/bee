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
