package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
)

type Field struct {
	Name    string
	Type    string
	Null    string
	Key     sql.NullString
	Default sql.NullString
	Extract sql.NullString
	Comment string
}

type Table struct {
	Name    string
	Comment string
	Fields  []*Field
}

const (
	getCommentSQL      = "SELECT COLUMN_NAME,column_comment FROM INFORMATION_SCHEMA.Columns WHERE table_name=? AND table_schema=?"
	getTableCommentSQL = "SELECT TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE table_name=? AND table_schema=?"
)

func Desc(db *sql.DB, tableName, dbName string) (*Table, error) {
	descSQL := fmt.Sprintf("DESC %s", tableName)
	rows, err := db.Query(descSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	table := new(Table)
	table.Name = tableName
	table.Fields = make([]*Field, 0)
	for rows.Next() {
		field := new(Field)
		err = rows.Scan(&field.Name, &field.Type,
			&field.Null, &field.Key, &field.Default,
			&field.Extract)
		if err != nil {
			return nil, err
		}
		table.Fields = append(table.Fields, field)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		err = setTableComment(db, table, dbName)
		wg.Done()

	}()

	go func() {
		err = setFieldsComments(db, table, dbName)
		wg.Done()

	}()
	wg.Wait()

	if err != nil {
		return nil, err
	}

	return table, nil
}

func setTableComment(db *sql.DB, table *Table, dbName string) error {
	rows, err := db.Query(getTableCommentSQL, table.Name, dbName)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&table.Comment)
		if err != nil {
			return err
		}
	}
	return nil
}

func setFieldsComments(db *sql.DB, table *Table, dbName string) error {
	rows, err := db.Query(getCommentSQL, table.Name, dbName)
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
		for _, f := range table.Fields {
			if f.Name == name {
				f.Comment = comment
				break
			}
		}
	}
	return nil
}

func ConvertSQLType(sqlType string) string {
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
		return "int8"
	case strings.HasPrefix(sqlType, "FLOAT"):
		return "float32"
	case strings.HasPrefix(sqlType, "DOUBLE"):
		return "float64"
	case strings.HasPrefix(sqlType, "DECIMAL"):
		return "float64"
	case strings.HasPrefix(sqlType, "DATE"):
		return "time.Time"
	case strings.HasPrefix(sqlType, "Time"):
		return "time.Time"
	}
	// 默认情况下为string
	return "string"
}
