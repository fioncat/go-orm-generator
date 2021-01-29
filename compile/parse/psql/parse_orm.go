package psql

import (
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/generate/coder"
	"github.com/fioncat/go-gendb/misc/errors"
)

// OrmResult represents the result of parsing the data
// table and is used to produce ORM code. It basically
// stores the basic information of the data table.
type OrmResult struct {
	// TableName represents the name of the table on the
	// database. It is generally in the form of underlining.
	TableName string

	// GoName is the name converted from TableName to
	// camel case, and the first letter is capitalized.
	GoName string

	// Struct represents the ORM structure that needs
	// to be generated.
	Struct coder.Struct

	// Fields represents the fields' slice contained in
	// the ORM structure.
	Fields []*OrmField

	// IdField represents id, and each ORM table must
	// meet: there is one and only one id. This field
	// stores the id field.
	IdField *OrmField

	// WithoutIdFields stores all non-id fields. It has
	// one element less than Fields.
	WithoutIdFields []*OrmField
}

func (r *OrmResult) Type() string {
	return "sql-orm"
}

func (r *OrmResult) Key() string {
	return "orm." + r.GoName
}

func (r *OrmResult) GetStructs() []*coder.Struct {
	return nil
}

// OrmField represents a specific field of the ORM structure.
type OrmField struct {
	Name   string
	GoName string
	GoType string

	IsPrimary  bool
	IsAutoIncr bool
}

// parse orm, need to connect database.
func orm(tableName string) (*OrmResult, error) {
	if err := rdb.MustInit(); err != nil {
		return nil, err
	}

	var table rdb.Table
	v, ok := tableCache.Load(tableName)
	if !ok {
		descTable, err := rdb.Get().Desc(tableName)
		if err != nil {
			return nil, errors.Trace("orm: desc table", err)
		}
		tableCache.Store(tableName, table)
		table = descTable
	} else {
		table = v.(rdb.Table)
	}

	result := new(OrmResult)
	result.TableName = tableName
	result.GoName = coder.GoName(tableName)

	result.Struct.Name = result.GoName
	result.Struct.Comment = table.GetComment()

	fieldNames := table.FieldNames()
	result.WithoutIdFields = make([]*OrmField, 0, len(fieldNames)-1)
	result.Fields = make([]*OrmField, len(fieldNames))
	result.Struct.Fields = make([]coder.Field, len(fieldNames))

	for idx, fieldName := range fieldNames {
		field := table.Field(fieldName)

		ormField := new(OrmField)
		ormField.Name = fieldName
		ormField.GoName = coder.GoName(fieldName)
		ormField.GoType = rdb.Get().GoType(field.GetType())
		ormField.IsPrimary = field.IsPrimaryKey()
		ormField.IsAutoIncr = field.IsAutoIncr()

		if ormField.IsPrimary {
			if result.IdField != nil {
				return nil, errors.New(`found multi primary keys `+
					`for table "%s", now we donot support orm-code`+
					` for it`, tableName)
			}
			result.IdField = ormField
		} else {
			result.WithoutIdFields = append(
				result.WithoutIdFields, ormField)
		}

		result.Fields[idx] = ormField

		var goField coder.Field
		goField.Name = ormField.GoName
		goField.Type = ormField.GoType
		goField.Comment = field.GetComment()
		goField.AddTag("field", fieldName)
		if ormField.IsPrimary {
			goField.AddTag("id", "true")
		}

		result.Struct.Fields[idx] = goField
	}

	if result.IdField == nil {
		return nil, errors.New(`missing primary key for table "%s"`, tableName)
	}

	return result, nil
}
