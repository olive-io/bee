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
