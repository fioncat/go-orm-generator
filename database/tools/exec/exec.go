package exec

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/sql"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/database/tools/common"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/misc/term"
)

type Arg struct {
	Cache bool `flag:"cache"`
	Log   bool `flag:"log"`

	DbType string `flag:"db-type" default:"mysql"`

	FmtType string `flag:"fmt-type" default:"fields"`

	RowsLimit int `flag:"rows-limit" default:"100"`

	Conn   string `arg:"conn"`
	Path   string `arg:"path"`
	Method string `arg:"method"`
}

const maxRowsLimit = 1000

func Do(arg *Arg) error {
	if arg.RowsLimit >= maxRowsLimit {
		return fmt.Errorf("rows limit is too big, max %d",
			maxRowsLimit)
	}
	if arg.RowsLimit <= 0 {
		return fmt.Errorf("rows limit must bigger than 0")
	}
	if arg.Cache {
		rdb.EnableTableCache = true
	}
	if arg.Log {
		log.Init(true, "")
	}
	err := rdb.Init(arg.Conn, arg.DbType)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(arg.Path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")

	compileErr := func(err error) error {
		return errors.OnCompile(arg.Path, lines, err)
	}

	file, err := sql.ReadLines(arg.Path, lines)
	if err != nil {
		return err
	}

	var execM *sql.Method
	for _, m := range file.Methods {
		if m.Name == arg.Method {
			execM = m
			break
		}
	}
	if execM == nil {
		return fmt.Errorf(`can not find method `+
			`"%s" in file %s`, arg.Method, arg.Path)
	}

	if !common.FindMethodTag(execM, "exec") {
		err := execM.FmtError("none exec params")
		return compileErr(err)
	}

	exec, err := common.Method2Exec(execM, "exec")
	if err != nil {
		return compileErr(err)
	}

	if execM.Exec {
		return execSql(exec)
	}

	qfs, err := buildQueryFields(execM)
	if err != nil {
		return err
	}

	return querySql(exec, arg, qfs)
}

func execSql(exec *common.Exec) error {
	start := time.Now()

	affect, err := rdb.Get().Exec(exec.Sql, exec.Vals...)
	if err != nil {
		return err
	}

	fmt.Printf("\nExec done, %d row(s) affected, took %v\n",
		affect, time.Since(start))
	return nil
}

type queryField struct {
	Name string
	Type string
}

func buildQueryFields(m *sql.Method) ([]*queryField, error) {
	qfs := make([]*queryField, len(m.Fields))
	for idx, f := range m.Fields {
		qf := new(queryField)
		qf.Name = f.Name
		if f.IsCount {
			qf.Type = "int64"
			qfs[idx] = qf
			continue
		}
		table, err := rdb.Get().Desc(f.Table)
		if err != nil {
			return nil, fmt.Errorf(`desc table "%s" failed: %v`,
				f.Table, err)
		}

		dbField := table.Field(f.Name)
		if dbField == nil {
			return nil, fmt.Errorf(`can not find field `+
				`"%s" in table "%s"`, f.Name, f.Table)
		}
		qf.Type = rdb.Get().GoType(dbField.GetType())

		qfs[idx] = qf
	}

	return qfs, nil
}

func querySql(exec *common.Exec, arg *Arg, qfs []*queryField) error {
	showFunc := showFmts[arg.FmtType]
	if showFunc == nil {
		return fmt.Errorf(`unknown fmt-type "%s"`, arg.FmtType)
	}
	start := time.Now()

	titles := make([]string, len(qfs))
	for idx, qf := range qfs {
		titles[idx] = qf.Name
	}

	rows, err := rdb.Get().Query(exec.Sql, exec.Vals...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var data [][]interface{}
	var total int64
	for rows.Next() {
		total += 1
		if len(data) >= arg.RowsLimit {
			continue
		}
		vals := make([]interface{}, len(qfs))
		for idx, f := range qfs {
			vals[idx] = newValue(f.Type)
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

	fmt.Printf("\ndone, totally had %d row(s), "+
		"displayed %d row(s), took %v\n",
		total, len(data), time.Since(start))
	return nil
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
