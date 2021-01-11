# Go-GenDB

go-gendb是一个快速生成数据库访问代码的工具。利用它可以将SQL语句定义在代码外部，通过特殊的标签将它们和Go代码中的方法或结构体进行关联，随后自动生成这些方法或结构体的实现。

这样做的好处是：

- SQL语句定义在外部sql文件中，编辑器直接原生支持高亮、自动补全等功能，方便编写语句。
- 在SQL语句中通过具名占位符"${name}"来引入参数，在代码生成时对占位符进行替换。使得参数更加清晰易懂，避免在SQL语句中写很多的"?"或"%s"这样的意义不明的占位符。
- （尚未支持）支持SQL语句复用，通过"@{name}"占位符引入，避免大量重复SQL的编写。
- （尚未支持）通过"!{if xxx} !{endif}"和"!{for xxx} !{endfor}"在SQL语句中插入动态内容，根据不同的参数来生成不同的SQL语句，避免字符串拼接操作的编写。
- 通过解析查询语句的SELECT子句，自动生成`rows.Scan(...)`，避免编写冗长的fields列表。
- 对于查询语句，支持连接数据库，读取查询字段的类型，自动生成查询语句的返回结构体。
- （尚未支持）对于特别简单的SQL语句，例如简单的CRUD，和简单的单表查询语句，支持直接生成SQL语句和Go调用函数。
- （尚未支持）对no-sql的支持
- ...

除此之外，go-gendb还提供了一些便捷的工具：

- check：一键调用数据库对编写的SQL语句进行"DESC"检查，发现语句语法错误和警告信息。例如查询语句的全表扫描。
- ...

希望它能给你带来快乐:)

## 安装

你可以到[release页面](https://github.com/fioncat/go-gendb/releases)直接下载二进制程序，并将它放到你的`PATH`中。

更推荐的做法是，直接在本机进行构建，整个过程很简单。go-gendb是用Go开发的，确保你的机器上有`go1.11`及以上的开发环境（需要开启`go module`），执行：

```text
$ go get github.com/fioncat/go-gendb
```

即可完成安装。输入`go-gendb version`可以查看安装的版本。

## Quick Start

下面以一个极简的例子展示go-gendb，它只包含了go-gendb最简单的特性，如果要查看完整特性，请参见[samples](samples)下的更多例子，下面这个Quick Start仅适用于想要快速预览go-gendb的人，Quck Start的例子可以在[samples/quickstart](samples/quickstart)下找到。

首先，假设我们手头有一个mysql数据库，上面有一个简单的表`user`:

```sql
CREATE TABLE `user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '用户id',
  `name` varchar(50) DEFAULT NULL COMMENT '用户名',
  `email` varchar(50) DEFAULT NULL COMMENT '用户邮箱',
  `phone` varchar(50) DEFAULT NULL COMMENT '用户电话',
  `age` int(11) DEFAULT NULL COMMENT '用户年龄',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `password` varchar(50) DEFAULT NULL COMMENT '用户密码',
  `is_admin` int(11) DEFAULT NULL COMMENT '是否是管理员 0-不是  1-是',
  `is_delete` int(8) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB;
```

我们想为这个表编写一些操作，包括：

- FindById: 根据id查询单个用户并返回
- FindAdmins: 查询所有的admin用户并返回
- UpdateAge: 根据id更新某个用户的年龄
- Delete: 根据id对用户进行逻辑删除
- Count: 统计所有用户的个数

首先，要为这些操作定义Go方法，我们只需要新建一个go文件`oper.go`，并编写以下的代码：

```go
// +gendb sql

package user

import "database/sql"

// +gendb user.sql
type UserDb interface {
    // +gendb auto-ret
    // 根据id查询单个用户并返回
    FindById(db *sql.DB, id int64) (*User, error)

    // 查询所有的admin用户并返回
    FindAdmins(db *sql.DB) ([]*User, error)

    // 根据id更新某个用户的年龄
    UpdateAge(db *sql.DB, id int64, age int32) (sql.Result, error)

    // 根据id对用户进行逻辑删除
    Delete(db *sql.DB, id int64) (sql.Result, error)

    // 统计所有用户的个数
    Count(db *sql.DB) (int64, error)
}
```

可见各种标记是通过注释引入的，go-gendb会识别这些注释，如果想要生成数据库操作方法，必须遵从：

- 代码的"package"定义之前必须有"+gendb sql"声明。
- 将要生成的函数放在接口里面，并通过"+gendb {path/to/sql}"来标记接口。其中后面的"{path/to/sql}"表示该接口关联的sql文件的路径（相对当前代码文件的相对路径），可以有多个（通过空格分隔），我们在稍后会定义这个sql文件。在上述例子中`UserDb`通过"+gendb user.sql"标记，说明它关联的SQL在代码路径下的"user.sql"中。
- 接口里面每个方法的定义必须满足："{name}({params}) ({ret}, error)"的形式。其中，参数部分和正常的Go函数参数定义一样，没有什么限制。但是返回值的第二个必须是`error`，第一个返回值对于查询语句来说，如果是非切片，则表示SQL语句返回一条记录，如果是切片则表示返回多条记录；如果语句是执行语句，必须是`sql.Result`或`int64`。
- 使用"auto-ret"标记查询方法，表示会自动为该方法生成返回结构体。如果你懒得定义方法的返回结构体可以加上，但是为了获取字段类型，go-gendb在生成代码中必须连接数据库。

定义好之后，需要定义上面的方法对应的SQL语句，我们已经通过"+gendb user.sql"标记了"user.sql"为该接口SQL语句的位置，则在当前目录下创建"user.sql"文件并编写SQL语句：

```sql
-- UserDb的SQL语句

-- !FindById
SELECT
	id, name, email, phone, age, create_time,
	password, is_admin, is_delete

FROM user
WHERE id=${id} AND is_delete=0;

-- !FindAdmins
SELECT
	id, name, email, phone, age, create_time,
 	password, is_admin, is_delete

FROM user
WHERE is_admin=1 AND is_delete=0;

-- !UpdateAge
UPDATE user SET age=${age}
WHERE id=${id};

-- !DeleteUser
UPDATE user SET is_delete=1
WHERE id=${id};

-- !Count
SELECT COUNT(1) FROM user WHERE is_delete=0;
```

可以看到，通过"-- !{name}"的形式将SQL语句和方法进行关联。在SQL语句中通过"${name}"表示参数。

因为上述过程涉及到`auto-gen`所以需要配置数据库连接，通过以下的命令配置：

```text
$ go-gendb conn-set test '127.0.0.1' --user 'test' --pass 'test' --db 'test'
```

我这是在本地mysql数据库创建的用户、密码、数据库名称都为"test"，其中包含了上述的"user"表。

如果你没有使用`auto-gen`，可以需要上述的配置。

完成之后，执行生成命令：

```text
$ go-gendb gen --conn test oper.go
```

即可完成代码生成，所有的代码会被生成到"oper.go"的同级目录之下。

生成的代码包括两个文件，其中，"zz_generated_Struct.go"保存了生成的"User"结构体，你可以到[samples/quickstart/zz_generated_Struct.go](samples/quickstart/zz_generated_Struct.go)下查看；"zz_generated_oper.UserDb.go"保存了"UserDb"接口的实现，可以到[samples/quickstart/zz_generated_oper.UserDb.go](samples/quickstart/zz_generated_oper.UserDb.go)下查看。

在使用的时候，可以直接用生成的"UserDbOper"全局变量实现操作，它实现了"UserDb"接口：

```go
package main

import (
  "database/sql"
  
  user "github.com/fioncat/go-gendb/samples/quickstart"
)

func main() {
  db, err := sql.Open("mysql", ...)
  if err != nil { ... }
  
  // 调用FindById
  user, err := user.UserDbOper.FindById(db, 1)
  // 调用FindAdmins
  users, err := user.UserDbOper.FindAdmins(db)
  // 调用UpdateAge
  _, err = user.UserDbOper.UpdateAge(db, 1, 23)
  // 调用DeleteUser
  _, err = user.UserDbOper.DeleteUser(db, 2)
  // 调用Count
  total, err := user.UserDbOper.Count(db)
}
```

可见，我们只需要关心方法的定义和SQL语句的编写，具体SQL是如何被调用的完全不需要关心。

go-gendb的功能远不止于此，更多功能请参见进阶例子。

