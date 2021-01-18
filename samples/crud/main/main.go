package main

import (
	"database/sql"
	"fmt"

	"github.com/fioncat/go-gendb/database/conn"
	"github.com/fioncat/go-gendb/misc/addr"
	"github.com/fioncat/go-gendb/misc/errors"
	user "github.com/fioncat/go-gendb/samples/crud"
	_ "github.com/go-sql-driver/mysql"
)

func mysqlConnect(conn *conn.Config) (*sql.DB, error) {
	addr, err := addr.Parse(conn.Addr, 3306)
	if err != nil {
		return nil, errors.Trace("parse addr", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=5s&readTimeout=6s",
		conn.User, conn.Password, addr.Host,
		addr.Port, conn.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Trace("connect db", err)
	}
	db.SetMaxOpenConns(20)
	return db, nil
}

func main() {
	conn := conn.Config{
		Addr:     "127.0.0.1:3306",
		User:     "test",
		Password: "test",
		Database: "test",
	}

	db, err := mysqlConnect(&conn)
	if err != nil {
		fmt.Printf("connect mysql failed: %v\n", err)
		return
	}

	u, err := user.UserORM.FindById(db, 110)
	if err != nil {
		fmt.Println(err)
		return
	}

	u.Name = "哈哈哈哈"
	_, err = u.Save(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(u)
}
