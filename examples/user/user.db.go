// +gendb mysql

package user

import (
	"database/sql"
)

// User ...
type User struct {
	// +gendb-struct user
	// +gendb-ignore is_delete
	// +gendb-type is_admin bool
	userORM
}

// +gendb-inter user.sql
type genUserOps interface {
	// +gendb-lastid
	// add user
	add(db *sql.DB, u *User) (int64, error)

	// +gendb-affect
	// update user
	update(db *sql.DB, u *User) (int64, error)

	findById(db *sql.DB, id int64) (*User, error)
	search(db *sql.DB, where string) ([]User, error)
	searchConds(db *sql.DB, email string, phone string) ([]*User, error)

	// +gendb-count
	count(db *sql.DB) (int32, error)

	// +gendb-count
	countAdmin(db *sql.DB, admin int32) (int64, error)
}
