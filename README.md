# go-gendb

## Introduction

go-gendb is a tool to generate Go database code. Its basic idea is to separate the database logic (such as SQL statement) and code. go-gendb will link them together.

## Install

There are already compiled binary progams on [release page](https://github.com/fioncat/go-gendb/releases) which can be downloaded and used directly. But it is more recommended to build and install locally:

```text
GO111MODULE=on go install github.com/fioncat/go-gendb
```

In `go 1.16` and later, you can use the following command to specify the version:

```text
go install github.com/fioncat/go-gendb@version
```

## Versions

### v0.2.x and before

Early version, you won't use it.

These versions were used for testing early, and they should be sealed in history instead of being used in a production environment. The usage of the earlier version is not compatible with the current version, and there may be many problems. I no longer intend to provide documentation for these versions, so no one but myself may be able to use it anymore.

### v0.3.x

This is the first version I used in the production environment. It has been confirmed to be able to generate some code. Of course, because there are too few people testing it, there may be many potential bugs.

It can support the following basic functions:

- Generate code to call sql statement (including `db.Query`, `db.Exec`, `rows.Scan`, `rows.Close` and other calls), users only need to care about interface definition and sql statement writing.
- Support inserting `${name}` and `#{name}` placeholders in sql statements to indicate `"?"` and `"%v"` parameters.
- Support inserting `%{if cond} ... %{endif}` and `%{for ele in slice join 'x'} ... %{endfor}` placeholders in sql statements to write dynamic sql statements . Sql statement splicing code will be automatically generated.
- Support sql reuse, define some common sql statements and introduce them through `@{name}`.
- Support deriving its return structure definition based on query statement. This feature needs to connect to the database (to obtain the type of field).

## Usage

TODO: Doc is missing.	

For usage examples, please refer to [samples](samples).

