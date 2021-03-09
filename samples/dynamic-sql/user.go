// +gen:sql v=0.3 conn=local db_use=getDb()

package user

import "database/sql"

func getDb() *sql.DB {
	return nil
}

// +gen:UserOper user.sql
type _oper interface {
	// +gen:auto-ret
	FindByCond(conds map[string]interface{}, offset, limit int32) ([]*User, error)

	Adds(us []*User) (int64, error)
}
