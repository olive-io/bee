# tengo 模块 - "trace"

用户定义脚本的输出环境，类型各种 log 库，支持功能
- 输出为 json 格式
- 支持多种级别的 level
- 支持同时输出多个目录 /dev/stdout, 本地文件
- 支持多个输出目标
- 支持 webhook

```golang
flag := import("flag")
```

## 支持的方法

- `debug(args ...Object)`: 输出 debug 级别日志。
- `info(args ...Object)`: 输出 info 级别日志。
- `warn(args ...Object)`: 输出 warn 级别日志。
- `error(args ...Object)`: 输出 error 级别日志。
- `set_level(level string) => error`: 设置日志级别，默认为 info。
- `fields(args ...trace.Field) => trace/error`: 添加 fields 输出，支持链式操作 trace.fields(fields).info()。
- `int(name string, value int) => trace.Field/error`: 返回 int 类型的 Field.
- `float(name string, value float) => trace.Field/error`: 返回 float 类型的 Field.
- `string(name string, value string) => trace.Field/error`: 返回 string 类型的 Field.
- `time(name string, value Time) => trace.Field/error`: 返回 Time 类型的 Field.
- `add_handler(writer string, level string, args ...Object) => error`: 添加新的输出目标
- `add_hook(url string, args ...trace.Field) => []string/error`: 添加 webhook

## 实战实例

```go
trace := import("trace")
trace.set_level("debug")

name := "world"
// 追加 output
trace.add_handler("_output/tmp.log", "info", trace.string("namespace", "trace"))
// 新增 webhook
trace.add_hook("http://127.0.0.1:50001/hook", trace.string("namespace", "bee"))


fields := trace.string("a", "b")
trace.fields(fields).debug("name=%s", name)
trace.fields(fields).info("name=%s", name)
```

执行脚本
```shell
$ tengo ./flag_example.tengo --name=tengo
# name=tengo, on=false, targets=["a", "b"], ints=[1, 2, 3], floats=[1, 2.1, 3.3]
```