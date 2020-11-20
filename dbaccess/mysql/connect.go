package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/fioncat/go-gendb/misc/addr"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/store"
)

func Connect(conn *store.Conn) (*sql.DB, error) {
	addr, err := addr.Parse(conn.Addr, 3306)
	if err != nil {
		return nil, errors.Trace("parse addr", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=5s&readTimeout=6s",
		conn.User, conn.Pass, addr.Host,
		addr.Port, conn.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Trace("connect db", err)
	}
	db.SetMaxOpenConns(20)
	return db, nil
}
