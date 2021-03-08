// +gen:sql v=0.3 conn=local

package user

import "database/sql"

// +gen:UserOper user.sql
type _userInter interface {
	// +gen:auto-ret
	FindById(db *sql.DB, id int64) (*User, error)
	Add(db *sql.DB, u *User) (sql.Result, error)
}
