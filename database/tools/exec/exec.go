package exec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fioncat/go-gendb/compile/parse/parsesql"
	"github.com/fioncat/go-gendb/compile/scan/scansql"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/generate/coder"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/term"
)

// Arg represents the parameter structure of the exec command.
type Arg struct {
	QueryFormat string `flag:"fmt" default:"fields"`
	Params      string `flag:"m"`
	DbType      string `flag:"db-type" default:"mysql"`
	RowLimit    int    `flag:"rows-limit" default:"100"`

	Conn   string `arg:"conn"`
	Path   string `arg:"path"`
	Method string `arg:"method"`
}

// Do executes the execution of the sql statement in a
// certain sql file.
func Do(arg *Arg) error {
	err := rdb.Init(arg.Conn, arg.DbType)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(arg.Path)
	if err != nil {
		return err
	}

	scanResult, err := scansql.Do(arg.Path, string(data))
	if err != nil {
		return err
	}

	var tar *scansql.Statement
	for _, stat := range scanResult.Statements {
		if stat.Name == arg.Method {
			tar = &stat
			break
		}
	}
	if tar == nil {
		return fmt.Errorf(`can not find method "%s" in file %s`,
			arg.Method, arg.Path)
	}

	if tar.IsDynamic {
		return fmt.Errorf(`method "%s" is a dynamic method,`+
			` exec command donot support it.`, arg.Method)
	}

	method, err := parsesql.Statement(tar)
	if err != nil {
		return err
	}

	ms, err := parseParams(arg.Params)
	if err != nil {
		return err
	}

	// replace placeholders
	var reps = make([]interface{}, len(method.SQL.Replaces))
	for idx, repName := range method.SQL.Replaces {
		repValue := ms[repName]
		if repValue == nil {
			repStr := term.Input(`please input replace param "%s"`, repName)
			repValue = repStr
		}
		reps[idx] = repValue
	}

	// prepare placeholders
	pres := make([]interface{}, len(method.SQL.Prepares))
	for idx, preName := range method.SQL.Prepares {
		preValue := ms[preName]
		if preValue == nil {
			preStr := term.Input(`please input prepare param "%s"`, preName)
			preValue = preStr
		}
		pres[idx] = preValue
	}

	sqlStr := fmt.Sprintf(method.SQL.Contant, reps...)

	if method.IsExec {
		return exec(sqlStr, pres)
	}

	// count statement
	// TODO: This is a temporary solution for the COUNT
	// statement. Changing the COUNT solution in the
	// future requires changing the code here.
	if len(method.QueryFields) == 1 &&
		method.QueryFields[0].IsCount {
		// COUNT
		return count(sqlStr, pres)
	}

	retStruct, err := parsesql.AutoRet(method)
	if err != nil {
		return errors.Trace("parse return type", err)
	}

	fields := newFields(retStruct, method.QueryFields)

	return query(sqlStr, pres, fields, arg.QueryFormat, arg.RowLimit)
}

func parseParams(param string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	if param == "" {
		return m, nil
	}
	err := json.Unmarshal([]byte(param), &m)
	if err != nil {
		return nil, errors.Trace("parse param", err)
	}
	return m, nil
}

func exec(sql string, vals []interface{}) error {
	start := time.Now()

	affect, err := rdb.Get().Exec(sql, vals...)
	if err != nil {
		return err
	}

	fmt.Printf("\nExec done, %d row(s) affected, took %v\n",
		affect, time.Since(start))
	return nil
}

type queryField struct {
	name   string
	goType string
}

func newFields(s *coder.Struct, qfs []parsesql.QueryField) []queryField {
	queryFields := make([]queryField, len(s.Fields))
	for idx, f := range s.Fields {
		qf := qfs[idx]
		name := qf.Alias
		if name == "" {
			name = qf.Field
		}

		queryFields[idx] = queryField{
			name:   name,
			goType: f.Type,
		}
	}
	return queryFields
}

func newValue(goType string) interface{} {
	switch goType {
	case "int":
		return new(int)
	case "int32":
		return new(int32)
	case "int64":
		return new(int64)
	case "float32":
		return new(float32)
	case "float64":
		return new(float64)
	}
	return new(string)
}

func query(sql string, vals []interface{}, fields []queryField, fmtType string, rowsLimit int) error {
	if rowsLimit <= 0 {
		return errors.New(`rows limit must > 0`)
	}
	showFunc := showFmts[fmtType]
	if showFunc == nil {
		return errors.New(`unknown fmt type "%s"`, fmtType)
	}

	start := time.Now()

	titles := make([]string, len(fields))
	for idx, f := range fields {
		titles[idx] = f.name
	}

	rows, err := rdb.Get().Query(sql, vals...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var data [][]interface{}
	var total int64
	for rows.Next() {
		total += 1
		if len(data) >= rowsLimit {
			continue
		}
		vals := make([]interface{}, len(fields))
		for idx, f := range fields {
			vals[idx] = newValue(f.goType)
		}
		err = rows.Scan(vals...)
		if err != nil {
			return errors.Trace("scan fields", err)
		}
		for idx, val := range vals {
			rv := reflect.ValueOf(val)
			vals[idx] = rv.Elem().Interface()
		}
		data = append(data, vals)
	}

	fmt.Println()
	if len(data) == 0 {
		fmt.Println("<empty set>")
	} else {
		showFunc(titles, data)
	}

	fmt.Printf("\ndone, totally had %d row(s), displayed %d row(s), took %v\n",
		total, len(data), time.Since(start))
	return nil
}

var showFmts = map[string]showQueryFunc{
	"table":  showTable,
	"json":   showJson,
	"fields": showFields,
}

type showQueryFunc func(titles []string, data [][]interface{})

func showTable(titles []string, data [][]interface{}) {
	maxLens := make([]int, len(titles))
	for idx, title := range titles {
		maxLens[idx] = len(title)
	}
	for _, row := range data {
		for idx, cell := range row {
			length := len(fmt.Sprint(cell))
			if maxLens[idx] < length {
				maxLens[idx] = length
			}
		}
	}

	fmtShow := func(idx int, val interface{}) string {
		len := maxLens[idx]
		layer := "%-" + strconv.Itoa(len) + "s"
		return fmt.Sprintf(layer, fmt.Sprint(val))
	}

	titleRow := make([]string, len(titles))
	for idx, title := range titles {
		titleRow[idx] = fmtShow(idx, title)
	}
	titleRowStr := strings.Join(titleRow, "|")
	titleRowLen := len(titleRowStr)

	fmt.Println(titleRowStr)
	fmt.Println(strings.Repeat("-", titleRowLen))

	for _, row := range data {
		dataRow := make([]string, len(row))
		for idx, cell := range row {
			dataRow[idx] = fmtShow(idx, cell)
		}
		fmt.Println(strings.Join(dataRow, "|"))
	}
}

func showJson(titles []string, data [][]interface{}) {
	for rowN, line := range data {
		m := make(map[string]interface{}, len(line))
		for idx, title := range titles {
			m[title] = line[idx]
		}
		fmt.Printf("row %d\n", rowN+1)
		term.Show(m)
		fmt.Println()
	}
}

func showFields(titles []string, data [][]interface{}) {
	for rowN, line := range data {
		fmt.Printf("row %d\n", rowN+1)
		for idx, title := range titles {
			val := line[idx]
			fmt.Printf("  %s: %s\n", title, value2str(val))
		}
		fmt.Println()
	}
}

func value2str(v interface{}) string {
	switch v.(type) {
	case string:
		return coder.Quote(v.(string))

	default:
		return fmt.Sprint(v)
	}
}

func count(sql string, vs []interface{}) error {
	start := time.Now()
	rows, err := rdb.Get().Query(sql, vs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cnt int64
	if rows.Next() {
		err = rows.Scan(&cnt)
		if err != nil {
			return errors.Trace("scan", err)
		}
	}

	fmt.Printf("count = %d, took %v\n", cnt, time.Since(start))
	return nil
}
