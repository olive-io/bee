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

package bee_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/olive-io/bee"
	pb "github.com/olive-io/bee/api/rpc"
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/parser"
	bs "github.com/olive-io/bee/server/grpc"
	"github.com/olive-io/bee/vars"
)

const hostText = `
host1 bee_host=192.168.2.141 bee_user=root bee_ssh_passwd=xxx
localhost bee_connect=grpc bee_host=127.0.0.1 bee_port=15250 bee_platform=linux bee_arch=amd64 bee_home=/tmp/bee
host2 bee_host=192.168.2.164 bee_connect=winrm bee_platform=windows bee_user=Administrator bee_home=C:\\Windows\\Temp\\bee 
`

func startGRPCServer(t *testing.T) {
	port := 15250
	addr := fmt.Sprintf("localhost:%d", port)

	ep := keepalive.EnforcementPolicy{
		MinTime:             time.Second * 30,
		PermitWithoutStream: true,
	}

	kp := keepalive.ServerParameters{
		//MaxConnectionIdle:     30 * time.Second,
		//MaxConnectionAge:      45 * time.Second,
		//MaxConnectionAgeGrace: 15 * time.Second,
		Time:    2 * time.Hour,
		Timeout: 20 * time.Second,
	}

	impl := bs.NewServer()
	server := grpc.NewServer(
		grpc.KeepaliveParams(kp), grpc.KeepaliveEnforcementPolicy(ep),
	)
	pb.RegisterRemoteRPCServer(server, impl)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_ = server.Serve(ln)
	}()

	time.Sleep(time.Second * 1)
}

func newRuntime(t *testing.T) (*bee.Runtime, *inv.Manager, func()) {
	startGRPCServer(t)

	dataloader := parser.NewDataLoader()
	if err := dataloader.ParseString(hostText); err != nil {
		t.Fatal(err)
	}
	inventory, err := inv.NewInventoryManager(dataloader)
	if err != nil {
		t.Fatal(err)
	}
	variables := vars.NewVariablesManager(dataloader, inventory)

	options := []bee.Option{
		bee.SetDir("_output/bee"),
		bee.SetCaller(func(ctx context.Context, host, action string, in []byte, opts ...bee.RunOption) ([]byte, error) {
			ropt := bee.RunOptions{}
			for _, opt := range opts {
				opt(&ropt)
			}
			t.Logf("handle service %v, msg = %s\n", ropt.Metadata, string(in))

			var m map[string]any
			_ = json.Unmarshal(in, &m)
			if _, ok := m["catch"]; ok {
				return nil, errors.New("catch error")
			}
			return []byte(fmt.Sprintf(`{"result": "ok"}`)), nil
		}),
	}
	rt, err := bee.NewRuntime(inventory, variables, dataloader, options...)
	if err != nil {
		t.Fatal(err)
	}
	cancel := func() {
		_ = rt.Stop()
	}
	return rt, inventory, cancel
}

func Test_Runtime(t *testing.T) {
	sources := []string{"host1", "localhost"}
	rt, inventory, cancel := newRuntime(t)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	options = append(options, bee.WithRunSync(true))
	options = append(options, bee.WithArgs(map[string]string{"hook": "http://127.0.0.1:5100", "a": "bbb"}))
	inventory.AddSources(sources...)
	data, err := rt.Execute(ctx, "host1", "oracle.test option=sys/oracle@orcl", options...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

func Test_Copy(t *testing.T) {
	sources := []string{"host1"}
	rt, inventory, cancel := newRuntime(t)
	defer cancel()

	dst := "/tmp/1.txt"
	defer os.Remove(dst)

	_ = os.WriteFile(dst, []byte("hello world"), os.ModePerm)
	shell := fmt.Sprintf("copy src=%s dst=/tmp/1.txt", dst)

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	inventory.AddSources(sources...)
	data, err := rt.Execute(ctx, "host1", shell, options...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))

	shell = fmt.Sprintf("fetch src=%s dst=/tmp/11.txt", dst)
	data, err = rt.Execute(ctx, "host1", shell, options...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

const moduleCfg = `name: "hello_world"
long: "this is a hello-world module"
authors:
  - lack

commands:
  - name: sub
    script: "hello_world.tengo"
    params:
      - name: name
        type: string
        default: "world"

    returns:
      - name: "message"
        type: "string"
root: hello_world
script: hello_world.tengo
`

const hello_worldTengo = `
os := import("os")
fmt := import("fmt")
text := import("text")

name := "world"
if len(os.args()) != 0 {
    flag := os.args()[2]
	name = text.trim_prefix(flag, "--name=")
}

fmt.printf("{\"message\": \"%s\"}\n", name)
`

func initDir(t *testing.T, root string) (string, func()) {
	err := os.MkdirAll(root, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(root, "bee.yml"), []byte(moduleCfg), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(root, "hello_world.tengo"), []byte(hello_worldTengo), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	return root, func() {
		//_ = os.RemoveAll(root)
	}
}

func Test_Runtime_extra_modules(t *testing.T) {
	extra := "_output/bee/modules/hello_world"
	_, clean := initDir(t, extra)
	defer clean()

	sources := []string{"host1"}
	rt, inventory, cancel := newRuntime(t)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	inventory.AddSources(sources...)
	data, err := rt.Execute(ctx, "host1", "hello_world.sub name=lack", options...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
