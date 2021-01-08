// +gendb sql

package user

import "database/sql"

// +gendb user.sql
type UserDb interface {

	// +gendb auto-ret
	GetById(db *sql.DB, id string) ([]*User, error)

	// +gendb auto-ret
	List(db *sql.DB, offset int, limit int) ([]*User, error)

	Update(db *sql.DB, u *User) (sql.Result, error)

	// +gendb auto-ret
	GetDetail(db *sql.DB, id string) (*UserDetail, error)
}
