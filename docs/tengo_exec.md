# tengo 模块 - "exec"

启动一个新的进程，类型标准库 `exec`

```golang
exec := import("exec")
```

## 支持的方法
- `command(name string, argv ...string) => Command/error`: 启动一个新的命令进程

## Command

- `combined_output() => bytes/error`: 执行命令并返回其标准输出和标准错误输出。
- `output() => bytes/error`: 执行命令并返回其标准输出。
- `run() => error`: 启动指定的命令并等待其完成。
- `start() => error`: 启动指定的命令，但不等待其完成。
- `wait() => error`: 等待命令退出，并等待从标准输入复制或从标准输出或标准错误输出复制的任何内容完成。
- `set_path(path string)`: 设置要运行的命令的路径。
- `set_dir(dir string)`: 设置进程的工作目录。
- `set_env(env [string])`: 设置进程的环境。

## 实战实例

```go
// flag_example.tengo
exec := import("exec")
fmt := import("fmt")

cmd := exec.command("/bin/bash", "-c", "ifconfig")

out := combined_output()
fmt.printf("%s", string(out))
```

执行脚本
```shell
$ tengo ./flag_example.tengo --name=tengo
# name=tengo, on=false, targets=["a", "b"], ints=[1, 2, 3], floats=[1, 2.1, 3.3]
```