// +gen:sql v=0.3

package user

import "database/sql"

// +gen:UserOper user.sql
type _userInter interface {
	FindById(db *sql.DB, id int64) (*User, error)
	Add(db *sql.DB, u *User) (sql.Result, error)
}
