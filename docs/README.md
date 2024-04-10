# tengo 手册

容灾项目中使用 [tengo](https://github.com/d5/tengo)  作为脚本库运行环境解释器。tengo 官方文档目前还不够完善(详看[此处](https://github.com/d5/tengo/blob/master/docs))，这再进行一些补充说明。

# 语法说明

## 基本语法

tengo 作为 Go 开发的解释型语言，语法和 Go 原生接近，基本上可以按照 Go 方式编写。详细的语法说明可以参考以下：

- 入门：https://github.com/d5/tengo/blob/master/docs/tutorial.md
- 内建函数：https://github.com/d5/tengo/blob/master/docs/builtins.md
- 数据类型：https://github.com/d5/tengo/blob/master/docs/runtime-types.md
- 运算符：https://github.com/d5/tengo/blob/master/docs/operators.md

> 注意：tengo 中不支持 `defer` 函数，脚本提前退出使用 `os.exit(1)` 方法。

## 函数

tengo 中支持自定义函数，但是需要注意使用方式。定义函数目前只支持 `匿名函数` 写法

```go
// 定义函数
sum := func(a, b) {
	return a + b
}

// 调用函数
sum(1, 2)
```

## 结构体

tengo 中的结构体定义方式类似于 Map，定义结构体目前只支持 `匿名结构体`写法

```go
fmt := import("fmt")

name := "tengo's struct"
count := 1

// 定义方法
print := func() {
	fmt.printf("name=%s, count=%d", name, count)
}

// 定义方法
set_name := func(text) {
	name = text
}

// 定义方法

// 定义结构体
custom := {
	name: name,
	count: count,
	print: print,
	set_name: set_name,
	set_count: func(c) {
		count = c
	}
}

// 使用结构体
new_name := "set " + custom.name
custom.set_name(new_name)
custom.print()
// Output: name=set tengo'struct, count=1
```

## 第三方库

tengo 使用 `import` 调用 `标准库` 和 `第三方库` 。

```go
fmt := import("fmt")
fmt.println("hello tengo")
```

这里说明如何自定义第三方库：

创建目录，结构如下：

```bash
# tree .
.
├── hello
│   ├── prefix.tengo
│   ├── hello.tengo
│   └── print.tengo
└── main.tengo
```

给 `hello` 库的 `hello.tengo` 添加全局变量，给 `print.tengo` 添加 `print` 函数。

```go
// prefix.tengo
prefix := "tengo"

// 使用 export 导出
export prefix
```

```go
// print.tengo
fmt := import("fmt")

export func(prefix, text) {
	fmt.printf("[%s] %s\n", prefix, text)
}
```

```bash
// hello.tengo
prefix := import("./prefix")
print := import("./print")

pre := prefix

hello := {
	prefix: pre,
	set_prefix: func(text) {
		pre = text
	},
	print: func(text) {
        print(pre, text)
    }
}

export hello
```

> 注意: `export` 只支持导出变量，变量的类型可以是内置类型、函数和结构体。

```go
// main.tengo
hello := import("./hello")

hello.print("tengo's text")
hello.set_prefix("mytengo")
hello.print("tengo's text")
// 调用 main.tengo
// tengo -import ./hello main.tengo
// Output: 
// [tengo] tengo's text
// [mytengo] tengo's text
```

`-import` 选项配置第三方库的根路径，可以理解为 `$GOPATH`

`import("./print")`  库的路径支持全路径和与 `-import` 的相对路径。

# 扩展标准库

可以通过自定义 tengo 解释器，扩展标准库的内容。目前新增三个标准库：

- [exec](https://github.com/olive-io/bee/blob/main/docs/tengo_exec.md)：支持执行本地命令
- [flag](https://github.com/olive-io/bee/blob/main/docs/tengo_flag.md)：解析命令行参数
- [trace](https://github.com/olive-io/bee/blob/main/docs/tengo_trace.md)：记录脚本运行过程中的日志，支持输出到本地日子文件和 webhook
- [filepath](https://github.com/olive-io/bee/blob/main/docs/tengo_filepath.md)：文件路径库