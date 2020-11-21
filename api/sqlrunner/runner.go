package sqlrunner

import (
	"database/sql"
	"errors"
	"fmt"
)

func Replaces(vs ...interface{}) []interface{} {
	return vs
}

type IDB interface {
	Query(sql string, vs ...interface{}) (*sql.Rows, error)
	Exec(sql string, vs ...interface{}) (sql.Result, error)
}

var ErrNotFound = errors.New("data not found")

type Runner struct {
	sql string
}

func New(sql string) *Runner {
	return &Runner{sql: sql}
}

func (r *Runner) query(db IDB, rs, vs []interface{}) (*sql.Rows, error) {
	sql := r.sql
	if len(rs) > 0 {
		sql = fmt.Sprintf(sql, rs...)
	}
	return db.Query(sql, vs...)
}

func (r *Runner) exec(db IDB, rs, vs []interface{}) (sql.Result, error) {
	sql := r.sql
	if len(rs) > 0 {
		sql = fmt.Sprintf(sql, rs...)
	}
	return db.Exec(sql, vs...)
}

func (r *Runner) Exec(db IDB, rs, vs []interface{}) (sql.Result, error) {
	return r.exec(db, rs, vs)
}

func (r *Runner) ExecAffect(db IDB, rs, vs []interface{}) (int64, error) {
	result, err := r.exec(db, rs, vs)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *Runner) ExecLastId(db IDB, rs, vs []interface{}) (int64, error) {
	result, err := r.exec(db, rs, vs)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

type ScanFunc func(rows *sql.Rows) error

func (r *Runner) QueryMany(db IDB, rs, vs []interface{}, scanFunc ScanFunc) error {
	rows, err := r.query(db, rs, vs)
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

func (r *Runner) QueryOne(db IDB, rs, vs []interface{}, scanFunc ScanFunc) error {
	rows, err := r.query(db, rs, vs)
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
