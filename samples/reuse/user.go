// +gendb sql

package user

import "database/sql"

type Cond struct {
	Key string
	Val interface{}
	Op  string
}

// +gendb user.sql
type UserDb interface {
	// +gendb auto-ret
	FindById(db *sql.DB, id int64) (*User, error)

	// +gendb auto-ret
	Detail(db *sql.DB, conds []Cond) ([]*User, error)

	Search(db *sql.DB, conds []Cond, offset int, limit int) ([]*User, error)

	// +gendb affect
	Adds(db *sql.DB, users []*User) (int64, error)
}
