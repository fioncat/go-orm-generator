// +gendb sql

package user

import "database/sql"

type User struct {
	// +gendb mysql user
	UserORM
}

// +gendb user.sql
type UserDB interface {
	Add(db *sql.DB, u *User) (int64, error)
	Update(db *sql.DB, u *User) (int64, error)
	FindByID(db *sql.DB, id int64) (*User, error)
	FindByName(db *sql.DB, name string) ([]*User, error)
	Search(db *sql.DB, where string, offset int32, limit int32) ([]*User, error)
	Count(db *sql.DB, where string) (int64, error)
}
