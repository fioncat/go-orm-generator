package gensql

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fioncat/go-gendb/generator"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/parser"
	"github.com/fioncat/go-gendb/scanner"
)

type sqlResult struct {
	name string

	imports map[string]string
	methods []sqlMethod
}

const (
	sqlQueryMany  = 1
	sqlQueryOne   = 2
	sqlCount      = 3
	sqlExecAffect = 4
	sqlExecLastid = 5
	sqlExecResult = 6
)

type sqlMethod struct {
	name     string
	sql      string
	prepares []string
	replaces []string

	sqlType int

	retType string
	rets    []string

	definition string
}

func (m *sqlMethod) isQuery() bool {
	return m.sqlType == sqlQueryOne ||
		m.sqlType == sqlQueryMany
}

func parseInter(task *generator.Task, it generator.TaskInterface) (*generator.File, error) {
	// Read all required sql files
	if len(it.Args) == 0 {
		err := errors.Fmt("missing sql files for interface %s", it.Name)
		return nil, errors.Line(task.Path, it.Line, err)
	}

	dir := filepath.Dir(task.Path)

	sqlFile := make(generator.SQLFile)
	for _, sqlPath := range it.Args {
		sqlPath = filepath.Join(dir, sqlPath)
		file, err := scanner.SQL(sqlPath)
		if err != nil {
			return nil, errors.Trace("read sql file failed", err)
		}
		sqlFile.Merge(file)
	}

	log.Infof("parsing methods for interface %s", it.Name)
	file := new(sqlResult)
	file.name = it.Name
	file.methods = make([]sqlMethod, len(it.Methods))
	file.imports = make(map[string]string)
	for i, taskMethod := range it.Methods {
		// The parameter must include db, indicating
		// a database connection.
		if !taskMethod.ParamNames.Exists("db") {
			err := errors.Fmt(`missing "db" param for method %s`,
				taskMethod.Name)
			return nil, errors.Line(task.Path, taskMethod.Line, err)
		}
		var method sqlMethod
		method.name = taskMethod.Name
		method.definition = taskMethod.String()

		// find the sql statement of this method
		sql, ok := sqlFile[taskMethod.Name]
		if !ok {
			err := errors.Fmt(`can not find sql name "%s",`+
				`please check your sql file(s)`, taskMethod.Name)
			return nil, errors.Line(task.Path, taskMethod.Line, err)
		}
		method.prepares = sql.Prepares
		method.replaces = sql.Replaces
		method.sql = sql.Content
		method.retType = taskMethod.ReturnType

		for _, opt := range taskMethod.Options {
			switch opt.Tag {
			case "affect":
				method.sqlType = sqlExecAffect

			case "lastid":
				method.sqlType = sqlExecLastid

			case "count":
				method.sqlType = sqlCount

			case "many":
				method.sqlType = sqlQueryMany

			case "one":
				method.sqlType = sqlQueryOne

			default:
				err := errors.Fmt(`unknown option "%s" for %s`,
					opt.Tag, taskMethod.Name)
				return nil, errors.Line(
					task.Path, taskMethod.Line, err)
			}
		}

		// At this time, if the type is 0, it means
		// that the user has not specified it, and
		// it is automatically inferred based on
		// the sql statement and the return value.
		if method.sqlType == 0 {
			// Determine whether to query or execute
			// according to the sql statement.
			if isSQLQuery(method.sql) {
				if isSQLCount(method.sql) {
					method.sqlType = sqlCount
				} else {
					// Determine whether to
					// query multiple or one
					// based on the return value
					if parser.HasL(taskMethod.ReturnType, "[]") {
						method.sqlType = sqlQueryMany
					} else {
						method.sqlType = sqlQueryOne
					}
				}

			} else {
				// If the return value is int,
				// it will return affect by default,
				// if it is result, it will return
				// sql.Result. Otherwise, an error
				// is reported.
				if parser.HasL(taskMethod.ReturnType, "int") {
					method.sqlType = sqlExecAffect
				} else if taskMethod.ReturnType == "sql.Result" {
					method.sqlType = sqlExecResult
					file.imports["sql"] = "database/sql"
				} else {
					err := errors.Fmt("unsupport return "+
						"type %s for exec sql, method=%s",
						taskMethod.ReturnType, taskMethod.Name)
					return nil, errors.Line(task.Path,
						taskMethod.Line, err)
				}
			}

		}

		// Add the import in the method to the file.
		for pkg := range taskMethod.Pkgs {
			var foundImp *generator.Import
			for _, imp := range task.Imports {
				if imp.Name == pkg {
					foundImp = &imp
					break
				}
			}
			if foundImp == nil {
				err := errors.Fmt(`can not find `+
					`package "%s" in method %s`,
					pkg, taskMethod.Name)
				return nil, errors.Line(task.Path,
					taskMethod.Line, err)
			}
			file.imports[foundImp.Name] = foundImp.Path
		}

		// If it is a query, the query field of
		// the sql statement must be parsed.
		if method.isQuery() {
			rets, err := parser.SQLQuery(method.sql)
			if err != nil {
				return nil, errors.Fmt(`sql "%s" `+
					`format error: %v`,
					taskMethod.Name, err)
			}
			method.rets = rets
		}
		log.Infof(" > method %s, type=%d, prepares=%v,"+
			" replaces=%v, rets=%v", method.name,
			method.sqlType, method.prepares,
			method.replaces, method.rets)

		file.methods[i] = method
	}

	fileName := fmt.Sprintf("zz_generated_sql_%s.go",
		file.name)
	path := filepath.Join(dir, fileName)
	return &generator.File{
		Path:   path,
		Result: file,
	}, nil
}

func isSQLQuery(sql string) bool {
	sql = strings.ToUpper(sql)
	sql = parser.TS(sql)
	return parser.HasL(sql, "SELECT")
}

func isSQLCount(sql string) bool {
	sql = strings.ToUpper(sql)
	sql = parser.TS(sql)
	return parser.HasL(sql, "SELECT COUNT")
}
