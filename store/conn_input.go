package store

import (
	"encoding/json"
	"fmt"

	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/term"
)

func InputConn() (*Conn, error) {
	var conn Conn
	conn.Key = term.Input("Please input conn key")
	if conn.Key == "" {
		return missInput("key")
	}
	conn.Addr = term.Input("Please input addr")
	if conn.Addr == "" {
		return missInput("addr")
	}
	conn.User = term.Input("Please input user")
	conn.Pass = term.Input("Please input password")
	conn.Database = term.Input("Please input database")
	return &conn, nil
}

func ShowConn(conn *Conn) {
	data, _ := json.MarshalIndent(conn, "", "  ")
	fmt.Println(string(data))
}

func missInput(name string) (*Conn, error) {
	return nil, errors.Fmt("%s can not be empty", name)
}
