// +gendb sql

package user

import "database/sql"

type User struct {
	Id      int64
	Name    string
	Age     int
	Phone   string
	IsAdmin bool
}

// +gendb user.sql
type UserDb interface {
	// +gendb affect
	Update(db *sql.DB, u *User) (int64, error)

	FindByIds(db *sql.DB, ids []string) ([]*User, error)

	// +gendb affect
	BatchInsert(db *sql.DB, users []*User) (int64, error)

	FindById(db *sql.DB, id int64) (*User, error)
}
