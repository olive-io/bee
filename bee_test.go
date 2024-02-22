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

	"github.com/olive-io/bee"
	pb "github.com/olive-io/bee/api/rpc"
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/parser"
	bs "github.com/olive-io/bee/server/grpc"
	"github.com/olive-io/bee/vars"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const hostText = `
host1 bee_host=192.168.2.32 bee_user=root bee_ssh_passwd=123456
localhost bee_connect=grpc bee_host=127.0.0.1 bee_port=15250 bee_platform=linux bee_arch=amd64 bee_home=/tmp/bee
# host2 bee_host=192.168.2.164 bee_connect=winrm bee_platform=windows bee_user=Administrator bee_winrm_passwd=xxx
`

func startGRPCServer(t *testing.T) {
	port := 15250
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

	time.Sleep(time.Second * 1)
}

func newRuntime(t *testing.T, modules ...string) (*bee.Runtime, *inv.Manager, func()) {
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
		bee.SetModulePath(modules),
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
	//options = append(options, bee.SetRunSync(true))
	inventory.AddSources(sources...)
	data, err := rt.Execute(ctx, "localhost", "ping", options...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))

	//data, err = rt.Execute(ctx, "localhost", "ping", options...)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Log(string(data))
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
script: "hello_world.tengo"
authors:
  - lack
params:
  - name: name
    type: string
    default: "world"

returns:
  - name: "message"
    type: "string"
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
	extra := "_output/mymodule/hello_world"
	_, clean := initDir(t, extra)
	defer clean()

	sources := []string{"host1"}
	rt, inventory, cancel := newRuntime(t, extra)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	inventory.AddSources(sources...)
	data, err := rt.Execute(ctx, "host1", "hello_world name=lack", options...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
