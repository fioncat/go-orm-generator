# go-gendb

[中文README](README_zh.md)

## Introduction

go-gendb is a tool to generate Go database code. Its basic idea is to separate the database logic (such as SQL statement) and code. go-gendb will link them together.

In addition, go-gendb also contains a tool chan for operating the database.

## Install

There are already compiled binary progams on [release page](https://github.com/fioncat/go-gendb/releases) which can be downloaded and used directly. But it is more recommended to build and install locally:

```text
GO111MODULE=on go install github.com/fioncat/go-gendb
```

In `go 1.16` and later, you can use the following command to specify the version:

```text
go install github.com/fioncat/go-gendb@version
```

## Version

### v0.0.x

Early version, you won't use it.

There is no documentation, the code is messy, maybe only I can use it.

### v0.1.x

The first version that can be used, but there are many bugs, it is not recommended to use it in a production environment.

- Support dynamic sql statement.
- Added tool chain.
- Code refactoring to optimize performance.

### v0.2.x

The purpose of this version is to make go-gendb available, and it is still under testing. This version is not compatible with `v0.1.x`.

- Optimize storage methods, concurrency, and speed up code generation.
- Fix bugs.

## Usage

TODO: Doc is missing.	

For usage examples, please refer to [samples](samples).

