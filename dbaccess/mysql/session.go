package mysql

import (
	"database/sql"

	"github.com/fioncat/go-gendb/dbaccess/dbtypes"
	"github.com/fioncat/go-gendb/store"
)

type Session struct {
	db       *sql.DB
	database string
}

func (s *Session) Connect(conn *store.Conn) error {
	db, err := Connect(conn)
	if err != nil {
		return err
	}
	s.db = db
	s.database = conn.Database
	return nil
}

func (s *Session) Desc(tableName string) (*dbtypes.Table, error) {
	table, err := Desc(s.db, tableName, s.database)
	if err != nil {
		return nil, err
	}
	descTable := new(dbtypes.Table)
	descTable.Name = table.Name
	descTable.Comment = table.Comment
	descTable.Fields = make([]dbtypes.Field, len(table.Fields))
	for i, f := range table.Fields {
		descField := new(dbtypes.Field)
		descField.Name = f.Name
		descField.Comment = f.Comment
		descField.Type = ConvertSQLType(f.Type)
		descTable.Fields[i] = *descField
	}
	return descTable, nil
}
