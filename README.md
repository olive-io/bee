# 简介
bee 是一套远程自动化命令行工具，支持在本地批量运行远程主机上的脚本。

支持的脚本格式：

- .tengo (windows | linux)
- .bat .ps (windows)
- .sh .bash (linux)

支持的远程连接协议

- ssh
- winrm
- grpc

# 实例

# 直接执行内置模块命令
```go
package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/olive-io/bee"
	inv "github.com/olive-io/bee/inventory"
	"github.com/olive-io/bee/parser"
	"github.com/olive-io/bee/plugins/callback"
	"github.com/olive-io/bee/vars"
)

var inventoryText = `
host1 bee_host=localhost:22 bee_user=root bee_ssh_passwd=123456
`

func main() {
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
		bee.SetDir(".bee"), // 该目录为 bee 根目录，存放内置模块 和 tengo 解释器
		bee.SetLogger(lg),
	}
	rt, err := bee.NewRuntime(inventory, variables, dataloader, &callback.BaseCallBack{}, options...)
	if err != nil {
		lg.Fatal("bee runtime", zap.Error(err))
	}

	ctx := context.TODO()
	execOptions := make([]bee.RunOption, 0)
	data, err := rt.Execute(ctx, "host1", "ping", execOptions...)
	if err != nil {
		lg.Fatal("bee runtime", zap.Error(err))
	}

	lg.Info("output", zap.String("data", string(data)))
}
```

bee 根目录结构
```text
_output/bee
├── db
├── modules
│   └── builtin
│       └── ping
│           ├── bee.yml
│           └── ping.tengo
└── repl
    ├── tengo.linux.amd64
    ├── tengo.linux.arm64
    └── tengo.windows.amd64.exe

```

## 自定义模块
模块结构
```text
_output/mymodule
└── hello_world
    ├── bee.yml
    └── hello_world.tengo
```
hello_world.tengo
```go
os := import("os")
fmt := import("fmt")
text := import("text")

name := "world"
if len(os.args()) != 0 {
    flag := os.args()[2]
	name = text.trim_prefix(flag, "--name=")
}

fmt.printf("{\"message\": \"%s\"}\n", name)
```

执行自定义模块命令
```go

var inventoryText = `
host1 bee_host=localhost:22 bee_user=root bee_ssh_passwd=123456
`

func main() {
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
		bee.SetDir(".bee"), // 该目录为 bee 根目录，存放内置模块 和 tengo 解释器
		bee.SetLogger(lg),
	}
	rt, err := bee.NewRuntime(inventory, variables, dataloader, &callback.BaseCallBack{}, options...)
	if err != nil {
		lg.Fatal("bee runtime", zap.Error(err))
	}

	ctx := context.TODO()
	execOptions := make([]bee.RunOption, 0)
	data, err := rt.Execute(ctx, "host1", "ping", execOptions...)
	if err != nil {
		lg.Fatal("bee runtime", zap.Error(err))
	}

	lg.Info("output", zap.String("data", string(data)))
}
```
