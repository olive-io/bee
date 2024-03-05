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

	"go.uber.org/zap"

	"github.com/olive-io/bee"
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/parser"
	"github.com/olive-io/bee/vars"
)

var inventoryText = `
host1 bee_host=localhost:22 bee_user=root bee_ssh_passwd=123456
`

func ExampleRuntime() {
	sources := []string{"host1"}

	lg, _ := zap.NewDevelopment()
	dataloader := parser.NewDataLoader()
	if err := dataloader.ParseString(inventoryText); err != nil {
		lg.Fatal("parse inventory", zap.Error(err))
	}
	inventory, err := inv.NewInventoryManager(dataloader, sources...)
	if err != nil {
		lg.Fatal("inventory manager", zap.Error(err))
	}
	variables := vars.NewVariablesManager(dataloader, inventory)

	options := []bee.Option{
		bee.SetDir("_output/bee"),
		bee.SetLogger(lg),
	}
	rt, err := bee.NewRuntime(inventory, variables, dataloader, options...)
	if err != nil {
		lg.Fatal("bee runtime", zap.Error(err))
	}

	ctx := context.TODO()
	execOptions := make([]bee.RunOption, 0)
	data, err := rt.Execute(ctx, "host1", "hello_world name=lack", execOptions...)
	if err != nil {
		lg.Fatal("bee runtime", zap.Error(err))
	}

	lg.Info("output", zap.String("data", string(data)))
}
