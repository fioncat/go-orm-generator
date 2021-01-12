// +gendb sql

package user

import "database/sql"

// +gendb user.sql
type UserDb interface {
	// +gendb auto-ret
	// 根据id查询单个用户并返回
	FindById(db *sql.DB, id int64) (*User, error)

	// 查询所有的admin用户并返回
	FindAdmins(db *sql.DB) ([]*User, error)

	// 根据id更新某个用户的年龄
	UpdateAge(db *sql.DB, id int64, age int32) (sql.Result, error)

	// 根据id对用户进行逻辑删除
	DeleteUser(db *sql.DB, id int64) (sql.Result, error)

	// 统计所有用户的个数
	Count(db *sql.DB) (int64, error)
}
