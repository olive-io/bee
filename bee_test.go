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
	"os"
	"path/filepath"
	"testing"

	"github.com/olive-io/bee"
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/parser"
	"github.com/olive-io/bee/plugins/callback"
	"github.com/olive-io/bee/vars"
)

const hostText = `
host1 bee_host=192.168.2.32:22 bee_user=root bee_ssh_passwd=123456
host2 bee_host=192.168.2.164 bee_connect=winrm bee_platform=windows bee_user=Administrator bee_winrm_passwd=xxx
`

func newRuntime(t *testing.T, sources []string, modules ...string) (*bee.Runtime, func()) {
	dataloader := parser.NewDataLoader()
	if err := dataloader.ParseString(hostText); err != nil {
		t.Fatal(err)
	}
	inventory, err := inv.NewInventoryManager(dataloader, sources...)
	if err != nil {
		t.Fatal(err)
	}
	variables := vars.NewVariablesManager(dataloader, inventory)

	options := []bee.Option{
		bee.SetDir("_output/bee"),
		bee.SetModulePath(modules),
	}
	rt, err := bee.NewRuntime(inventory, variables, dataloader, &callback.BaseCallBack{}, options...)
	if err != nil {
		t.Fatal(err)
	}
	cancel := func() {
		_ = rt.Stop()
	}
	return rt, cancel
}

func Test_Runtime(t *testing.T) {
	sources := []string{"host1"}
	rt, cancel := newRuntime(t, sources)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	data, err := rt.Execute(ctx, "host1", "ping", options...)
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
	rt, cancel := newRuntime(t, sources, extra)
	defer cancel()

	ctx := context.TODO()
	options := make([]bee.RunOption, 0)
	data, err := rt.Execute(ctx, "host1", "hello_world name=lack", options...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}