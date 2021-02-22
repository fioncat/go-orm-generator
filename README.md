# go-gendb

go-gendb是一个Go的数据库代码自动生成工具。它的基本思想是将数据库逻辑和代码抽离，分开维护，由go-gendb本身将二者关联起来生成代码。

除此之外，go-gendb还包含了一些操作数据库的简易工具。

## 安装

[release页面](releases)已经有编译好的二进制程序了，但是更加建议直接在本地构建安装：

```text
$ GO111MODULE=on go install github.com/fioncat/go-gendb
```

在`go1.16`及以后的版本，可以使用以下命令来指定版本（`"latest"`为最新版本）：

```text
$ go install github.com/fioncat/go-gendb@version
```

如果你的go不支持`go module`，可以`clone`项目到本地让后执行`go install`。

## 版本

### v0.0.x

早期版本，不建议使用。

### v0.1.x

第一个可以使用的版本，但是bug比较多，不建议在生产环境中使用。

- 支持动态sql语句
- 增加了多个工具链
- 代码重构，优化性能

### v0.2.x

该版本的目的主要是让go-gendb可用，目前还在测试中。该版本和`v0.1.x`不兼容。

- 优化储存方式，并发，加快代码生成的速度。
- 修复bug

## 使用

TODO: 文档暂缺

使用范例可以参见[samples](samples)。

