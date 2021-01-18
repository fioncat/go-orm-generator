package run

import (
	"database/sql"
	"errors"
	"fmt"
)

type IDB interface {
	Query(sql string, vs ...interface{}) (*sql.Rows, error)
	Exec(sql string, vs ...interface{}) (sql.Result, error)
}

var ErrNotFound = errors.New("data not found")

func query(db IDB, sql string, rs, vs []interface{}) (*sql.Rows, error) {
	if len(rs) > 0 {
		sql = fmt.Sprintf(sql, rs...)
	}
	return db.Query(sql, vs...)
}

func exec(db IDB, sql string, rs, vs []interface{}) (sql.Result, error) {
	if len(rs) > 0 {
		sql = fmt.Sprintf(sql, rs...)
	}
	return db.Exec(sql, vs...)
}

func Exec(db IDB, sql string, rs, vs []interface{}) (sql.Result, error) {
	return exec(db, sql, rs, vs)
}

func ExecAffect(db IDB, sql string, rs, vs []interface{}) (int64, error) {
	result, err := exec(db, sql, rs, vs)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func ExecLastId(db IDB, sql string, rs, vs []interface{}) (int64, error) {
	result, err := exec(db, sql, rs, vs)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

type ScanFunc func(rows *sql.Rows) error

func QueryMany(db IDB, sql string, rs, vs []interface{}, scanFunc ScanFunc) error {
	rows, err := query(db, sql, rs, vs)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err = scanFunc(rows)
		if err != nil {
			return err
		}
	}
	return nil
}

func QueryOne(db IDB, sql string, rs, vs []interface{}, scanFunc ScanFunc) error {
	rows, err := query(db, sql, rs, vs)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		err = scanFunc(rows)
		if err != nil {
			return err
		}
		return nil
	}
	return ErrNotFound
}
