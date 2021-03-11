// +gen:sql-orm v=0.3
// +gen:sql-orm conn=local
// +gen:sql-orm db=test create_sql=create.sql

package user

// 用户表
// +gen:orm table=user name="User"
// +gen:orm index=[Name,Phone],[CreateDate]
type _user struct {
	// +gen:orm flags=[auto-incr,primary]
	Id int64 // 用户自增id

	Name string // 用户名称

	Phone string // 用户电话号码

	// +gen:orm flags=[unique]
	Code string // 用户的唯一编码

	// +gen:orm type='tinyint(2)'
	// +gen:orm name="is_removed"
	IsDelete bool // 用户是否被删除

	CreateDate int64 // 用户创建时间
}
