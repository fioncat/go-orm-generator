// Code generated by go-gendb, DO NOT EDIT.
// go-gendb version: 0.3
// source: samples/quick-start/user.go
package user

import (
	run "github.com/fioncat/go-gendb/api/sql/run"
	sql "database/sql"
)

// all sql statement(s) to use
const (
	_UserOper_FindById = "SELECT id, name, age, FROM user WHERE id=?"
	_UserOper_Add      = "INSERT INTO user(id, name, age) VALUES (?, ?, ?)"
)

var UserOper = &_UserOper{}

// User is a struct auto generated by UserOper.FindById
type User struct {
	Id   int64  `table:"user" field:"id"`
	Name string `table:"user" field:"name"`
	Age  int32  `table:"user" field:"age"`
}

type _UserOper struct {}

func (*_UserOper) FindById(db *sql.DB, id int64) (*User, error) {
	var o *User
	err := run.QueryOne(db, _UserOper_FindById, nil, []interface{}{id}, func(rows *sql.Rows) error {
		o = new(User)
		return rows.Scan(&o.Id, &o.Name, &o.Age)
	})
	return o, err
}

func (*_UserOper) Add(db *sql.DB, u *User) (sql.Result, error) {
	return run.Exec(db, _UserOper_Add, nil, []interface{}{u.Id, u.Name, u.Age})
}