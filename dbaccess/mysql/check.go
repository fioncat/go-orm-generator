package mysql

import (
	"database/sql"
	"fmt"

	"github.com/fioncat/go-gendb/dbaccess/dbtypes"
)

type descResult struct {
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

func desc(db *sql.DB, sql string, prepares []interface{}) ([]*descResult, error) {
	sql = "DESC " + sql
	rows, err := db.Query(sql, prepares...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rs []*descResult
	for rows.Next() {
		r := new(descResult)
		err = rows.Scan(&r.Id, &r.SelectType, &r.Table, &r.Partotions,
			&r.Type, &r.PossibleKeys, &r.Key, &r.KeyLen, &r.Ref,
			&r.Rows, &r.Filtered, &r.Extra)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, nil
}

func checkResult(rs []*descResult) *dbtypes.CheckResult {
	cr := new(dbtypes.CheckResult)
	for _, r := range rs {
		if r.Type.String == "ALL" {
			cr.Warns = append(cr.Warns, fullTableScan(r))
		}
	}
	return cr
}

func fullTableScan(r *descResult) string {
	return fmt.Sprintf(`full-table-scan for table "%s", rows=%d`,
		r.Table.String, r.Rows.Int64)
}
