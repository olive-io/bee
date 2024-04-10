# tengo 模块 - "filepath"

操作文件路径，类似标准库 `filepath`

```golang
filepath := import("filepath")
```

## 内置常量
- `separator`: 文件路径分隔符 windows 为`/`, 类 unix 为 `\`
- `list_separator`: 文件列表分隔库 windows 为`;`, 类 unix 为 `:`
- `line_separator`: 文件换行符 windows 为`\r\n`, 类 unix 为 `\n`

## 支持的方法
- `clean(name string) => string`: 整理文件路径 `filepath.Clean()`。
- `is_local(name string) => bool`: 判断是否为本地路径。
- `to_slash(name string) => string`: 替换文件中的 / 为系统适配符号。
- `from_slash(name string) => string`: 替换文件中的系统适配符号为 /。
- `split_list(name string) => []string`: 切分 $PATH 路径。
- `split(name string) => []string`: 切分文件路径。
- `rel(base, target string) => string/error`: 获取两个路径的相关路径。
- `join(name ...string) => string`: 合并成一个文件路径。
- `abs(name string) => string`: 获取文件路径的绝对路径。
- `ext(name string) => string`: 获取文件路径的扩展。
- `base(name string) => string`: 获取文件路径中的文件名。
- `dir(name string) => string`: 获取文件路径中的目录。
- `volume_name(name string) => string`: 获取文件路径中的磁盘名称

## 实战实例

```go
// flag_example.tengo
filepath := import("filepath")
fmt := import("fmt")

a := "/opt"
b := "tengo/test.tengo"
fmt.println(filepath.rel(a, b))
fmt.println(filepath.join(a, b))
fmt.println(filepath.ext(b))
fmt.println(filepath.base(b))
```

