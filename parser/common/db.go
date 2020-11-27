package common

var dbType = "mysql"

func SetDBType(t string) {
	dbType = t
}

func GetDBType() string {
	return dbType
}
