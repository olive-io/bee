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
	"testing"

	"github.com/olive-io/bee"
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/parser"
	"github.com/olive-io/bee/playbook"
	"github.com/olive-io/bee/vars"
)

const hostText = `
host1 bee_host=192.168.2.32:22 bee_user=root bee_ssh_passwd=123456
host2 bee_host=192.168.2.164 bee_connect=winrm bee_platform=windows bee_user=Administrator bee_winrm_passwd=Howlink@1401
`

func newRuntime(t *testing.T, sources []string) (*bee.Runtime, func()) {
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
	}
	rt, err := bee.NewRuntime(inventory, variables, dataloader, options...)
	if err != nil {
		t.Fatal(err)
	}
	cancel := func() {
		_ = rt.Stop()
	}
	return rt, cancel
}

func Test_Runtime(t *testing.T) {
	sources := []string{"host1", "host2"}
	rt, cancel := newRuntime(t, sources)
	defer cancel()

	ctx := context.TODO()
	task := &playbook.Task{
		Name:   "my first task",
		Module: "ping",
		Hosts:  sources,
	}
	options := make([]bee.RunOption, 0)
	err := rt.Run(ctx, task, options...)
	if err != nil {
		t.Fatal(err)
	}
}
