# Go-GenDB

go-gendb is a tool for quickly generating Go database access code. Through it, we can define complex sql statements externally (in sql file) and define the interface for executing these sql statements in the Go code. go-gendb can automatically generate the implementation of this interface to realize the execution of sql.

go-gendb can even connect to a remote database, and automatically generate the corresponding Go structure and simple CRUD code for a table in the database.

go-gendb can bring the following conveniences for Go developers to write database code:

- It is no longer necessary to manually define a mapping structure for a table and their basic `ORM` operation codes (for example, insert one record, query based on Id, etc.).
- For complex queries (such as associated queries, sub-queries), it can be defined in an external sql file. Reference in the form of interfaces in Go code. Such references will be created automatically. You can only care about the definition of the interface and the writing of sql statements. The calling process does not involve reflection, and the efficiency is almost the same as writing code by yourself.
- For query statements, it is no longer necessary to write a lot of repetitive and mechanical `rows.Scan` code. go-gendb will automatically parse the fields in the SELECT clause in the SQL statement, and the Scan deserialization code will be automatically generated!
- Generate code by using tags (`+gendb`) in the Go code. No additional configuration files are involved, easy to use and lightweight.
- Support some more advanced SQL statement functions, such as dynamic SQL statement, SQL statement reuse, etc.
- Support to generate a variety of database codes, such as MySQL, MongoDB, etc. And support the generation of L2 cache code, including remote cache (usually Redis), local LRU cache. And can ensure data consistency through distributed locks based on different implementations (see the advanced usage document for details).

## Install

go-gendb is a command line program that runs in the terminal. Support darwin, Linux, Windows system. You can download the binary program of the corresponding system directly from the release page.

A better installation method is to clone the code directly to the local build and install. This method requires at least `Go 1.11` or higher on your machine:

```text
$ git clone https://github.com/fioncat/go-gendb
$ cd go-gendb
$ export GO111MODULE=on    # This command is to open go module, if it is already opened, can be ignored.
$ go install
```

If the installation goes well, execute the following command, you should see the version of go-gendb:

```text
$ go-gendb version
```

## Quick Start



