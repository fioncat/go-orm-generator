package scanner

import (
	"io/ioutil"
	"strings"

	"github.com/fioncat/go-gendb/generator"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/parser"
)

func SQL(path string) (generator.SQLFile, error) {
	log.Infof("Begin to scan sql file: %s", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	file := make(generator.SQLFile)
	var sql *generator.SQL
	contents := make([]string, 0)
	saveSql := func() error {
		if len(contents) == 0 {
			err := errors.Fmt("empty sql for %s", sql.Name)
			return errors.Line(path, sql.Line, err)
		}
		sql.Content = strings.Join(contents, " ")
		if _, ok := file[sql.Name]; ok {
			err := errors.Fmt("sql define duplcate: %s", sql.Name)
			return errors.Line(path, sql.Line, err)
		}
		contents = make([]string, 0)
		file[sql.Name] = *sql
		return nil
	}

	for i, line := range lines {
		line = parser.TS(line)
		if line == "" {
			continue
		}
		if parser.HasL(line, "-- !") {
			name := parser.TS(parser.TL(line, "-- !"))
			if name == "" {
				err := errors.New("empty name for sql")
				return nil, errors.Line(path, i+1, err)
			}
			// Save previous sql content
			if sql != nil {
				err := saveSql()
				if err != nil {
					return nil, err
				}
			}
			sql = new(generator.SQL)
			sql.Name = name
			sql.Line = i + 1
			continue
		}

		if parser.HasL(line, "--") {
			continue
		}

		contents = append(contents, line)
	}
	if sql != nil {
		err := saveSql()
		if err != nil {
			return nil, err
		}
	}
	log.Infof("scanned %d sql contents", len(file))
	for name, sql := range file {
		sql.Prepares, sql.Replaces, sql.Content =
			parser.SQLParam(sql.Content)
		log.Infof("  # %s: prepares=%v replaces=%v",
			name, sql.Prepares, sql.Replaces)
		sql.Content = strings.Join(
			strings.Fields(sql.Content), " ")
		file[name] = sql
	}

	return file, nil
}
