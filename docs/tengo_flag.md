# tengo 模块 - "flag"

用于解析脚本输入参数，类似 golang 的 cobra 库

```golang
flag := import("flag")
```

## 支持的方法

- `parse() => error`: 解析脚本参数，参数来自 `os.args()`。
- `int(name string, value int64, usage string) => int64/error`: 添加 int 类型参数。
- `int_array(name string, value []int64, usage string) => []int64, error`: 添加 []int64 类型 参数。
- `float(name string, value float64, usage string) => float64/error`: 添加 float64 类型参数.
- `float_array(name string, value []float64, usage string) => []float64/error`: 添加 []float 类型参数.
- `string(name string, value string, usage string) => string/error`: 添加 string 类型参数
- `string_array(name string, value string, usage string) => []string/error`: 添加 []string 类型参数
- `bool(name string, value bool, usage string) => bool/error`: 添加 bool 类型参数
- `get_int(name string) => int64/error`: 获取 int 类型参数
- `get_int_array(name string) => []int64/error`: 获取 []int 类型参数
- `get_float(name string) => float64/error`: 获取 float 类型参数
- `get_float_array(name string) => []float64/error`: 获取 []float 类型参数
- `get_string(name string) => string/error`: 获取 string 类型参数
- `get_string_array(name string) => []string/error`: 获取 []string 类型参数
- `get_bool(name string) => bool/error`: 获取 bool 类型参数

## 实战实例

```go
// flag_example.tengo
flag := import("flag")
fmt := import("fmt")

name := flag.string("name", "world", "名称")
on := flag.bool("enable", false, "是否开启")
targets := flag.string_array("targets", ["a", "b"], "")
ints := flag.int_array("ints", [1, 2, 3], "")
floats := flag.float_array("fs", [1.0, 2.1, 3.3], "")
flag.parse()

fmt.printf("name=%s, on=%v, targets=%v, ints=%v, floats=%v\n", name, on, targets, ints, floats)
```

执行脚本
```shell
$ tengo ./flag_example.tengo --name=tengo
# name=tengo, on=false, targets=["a", "b"], ints=[1, 2, 3], floats=[1, 2.1, 3.3]
```