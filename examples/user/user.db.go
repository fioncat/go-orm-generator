// +gendb mysql

package user

import (
	"database/sql"
	"fmt"
)

// User ...
type User struct {
	// +gendb-struct user
	*userORM
}

type Detail struct {
	// +gendb-struct user_detail
	*detailORM
}

// +gendb-inter sql/user.sql
type genUserOps interface {
	// +gendb-affect
	// add user
	add(db *sql.DB, u *User, w fmt.Formatter) (int64, error)

	// +gendb-lastid
	// update user
	update(db *sql.DB, u *User) (int64, error)

	// find user by id
	findById(db *sql.DB, id int64) (*User, error)

	// find all user
	all(db *sql.DB) ([]*User, error)

	// +gendb-count
	// count user
	countUser(db *sql.DB) (int64, error)
}

// +gendb-inter sql/user_detail.sql
type genDetailOps interface {
	add(db *sql.DB, u *Detail) (int64, error)

	findById(db *sql.DB, id int64) (*Detail, error)
}

func AddUser() {

}
