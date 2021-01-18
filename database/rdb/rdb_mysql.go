package rdb

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/fioncat/go-gendb/database/conn"
	"github.com/fioncat/go-gendb/misc/addr"
	"github.com/fioncat/go-gendb/misc/errors"
)

// implement of mysql session

func mysqlConnect(conn *conn.Config) (*sql.DB, error) {
	addr, err := addr.Parse(conn.Addr, 3306)
	if err != nil {
		return nil, errors.Trace("parse addr", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=5s&readTimeout=6s",
		conn.User, conn.Password, addr.Host,
		addr.Port, conn.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Trace("connect db", err)
	}
	db.SetMaxOpenConns(20)
	return db, nil
}

type mysqlField struct {
	Name    string
	Type    string
	Null    string
	Key     sql.NullString
	Default sql.NullString
	Extract sql.NullString
	Comment string
}

func (f *mysqlField) GetName() string    { return f.Name }
func (f *mysqlField) GetComment() string { return f.Comment }
func (f *mysqlField) GetType() string    { return f.Type }
func (f *mysqlField) IsPrimaryKey() bool { return f.Key.String == "PRI" }
func (f *mysqlField) IsAutoIncr() bool   { return f.Extract.String == "auto_increment" }

type mysqlTable struct {
	Name    string
	Comment string

	fields     map[string]*mysqlField
	fieldNames []string
}

func (t *mysqlTable) GetName() string         { return t.Name }
func (t *mysqlTable) GetComment() string      { return t.Comment }
func (t *mysqlTable) Field(name string) Field { return t.fields[name] }
func (t *mysqlTable) FieldNames() []string    { return t.fieldNames }

type mysqlCheckField struct {
	Id           sql.NullInt32
	SelectType   sql.NullString
	Table        sql.NullString
	Partotions   sql.NullString
	Type         sql.NullString
	PossibleKeys sql.NullString
	Key          sql.NullString
	KeyLen       sql.NullInt32
	Ref          sql.NullString
	Rows         sql.NullInt64
	Filtered     sql.NullString
	Extra        sql.NullString
}

type mysqlCheckResult struct {
	err   error
	warns []string
}

func (r *mysqlCheckResult) GetErr() error      { return r.err }
func (r *mysqlCheckResult) GetWarns() []string { return r.warns }

// Set comments to the table, which requires separate
// execution of additional SQL statements.
func (t *mysqlTable) setComment(db *sql.DB, dbName string) error {
	rows, err := db.Query(mysqlGetTableCommentSQL, t.Name, dbName)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&t.Comment)
		if err != nil {
			return err
		}
		if strings.Contains(t.Comment, "\n") {
			t.Comment = strings.ReplaceAll(t.Comment, "\n", " ")
		}
	}
	return nil
}

// Set comments to the fields, which requires separate
// execution of additional SQL statements.
func (t *mysqlTable) setFieldsComment(db *sql.DB, dbName string) error {
	rows, err := db.Query(mysqlGetCommentSQL, t.Name, dbName)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name    string
			comment string
		)
		err = rows.Scan(&name, &comment)
		if err != nil {
			return err
		}
		if strings.Contains(comment, "\n") {
			comment = strings.ReplaceAll(comment, "\n", " ")
		}
		field := t.fields[name]
		if field != nil {
			field.Comment = comment
		}
	}
	return nil
}

type mysqlOper struct {
	dbName string
}

const (
	mysqlGetCommentSQL      = "SELECT COLUMN_NAME,column_comment FROM INFORMATION_SCHEMA.Columns WHERE table_name=? AND table_schema=?"
	mysqlGetTableCommentSQL = "SELECT TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE table_name=? AND table_schema=?"
)

func (o *mysqlOper) Init(sess *Session) { o.dbName = sess.cfg.Database }

func (o *mysqlOper) Desc(db *sql.DB, tableName string) (Table, error) {
	sql := fmt.Sprintf("DESC %s", tableName)

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	table := new(mysqlTable)
	table.Name = tableName
	table.fields = make(map[string]*mysqlField)
	table.fieldNames = make([]string, 0)

	for rows.Next() {
		field := new(mysqlField)
		err = rows.Scan(&field.Name, &field.Type,
			&field.Null, &field.Key, &field.Default,
			&field.Extract)
		if err != nil {
			return nil, err
		}
		table.fields[field.Name] = field
		table.fieldNames = append(table.fieldNames, field.Name)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		err = table.setComment(db, o.dbName)
		wg.Done()
	}()

	go func() {
		err = table.setFieldsComment(db, o.dbName)
		wg.Done()
	}()
	wg.Wait()

	if err != nil {
		return nil, err
	}

	return table, nil
}

func (*mysqlOper) Check(db *sql.DB, sql string, prepares []interface{}) (CheckResult, error) {
	sql = "DESC " + sql
	result := new(mysqlCheckResult)
	rows, err := db.Query(sql, prepares...)
	if err != nil {
		result.err = err
		return result, nil
	}
	defer rows.Close()

	var fields []*mysqlCheckField
	for rows.Next() {
		r := new(mysqlCheckField)
		err = rows.Scan(&r.Id, &r.SelectType, &r.Table, &r.Partotions,
			&r.Type, &r.PossibleKeys, &r.Key, &r.KeyLen, &r.Ref,
			&r.Rows, &r.Filtered, &r.Extra)
		if err != nil {
			return nil, err
		}
		fields = append(fields, r)
	}

	for _, f := range fields {
		if f.Type.String == "ALL" {
			warn := fmt.Sprintf(`full-table-scan for table "%s", rows=%d`,
				f.Table.String, f.Rows.Int64)
			result.warns = append(result.warns, warn)
		}
	}

	return result, nil
}

func (*mysqlOper) ConvertType(sqlType string) string {
	sqlType = strings.ToUpper(sqlType)
	switch {
	case strings.HasPrefix(sqlType, "VARCHAR"):
		fallthrough
	case strings.HasPrefix(sqlType, "CHAR"):
		fallthrough
	case strings.HasPrefix(sqlType, "TEXT"):
		return "string"
	case strings.HasPrefix(sqlType, "INT"):
		return "int32"
	case strings.HasPrefix(sqlType, "BIGINT"):
		return "int64"
	case strings.HasPrefix(sqlType, "SMALLINT"):
		fallthrough
	case strings.HasPrefix(sqlType, "TINYINT"):
		return "int32"
	case strings.HasPrefix(sqlType, "FLOAT"):
		return "float32"
	case strings.HasPrefix(sqlType, "DOUBLE"):
		return "float64"
	case strings.HasPrefix(sqlType, "DECIMAL"):
		return "float64"
	case strings.HasPrefix(sqlType, "DATE"):
		return "string"
	case strings.HasPrefix(sqlType, "Time"):
		return "string"
	}
	return "string"
}
