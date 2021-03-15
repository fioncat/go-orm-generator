package generate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/compile/sql"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/version"
)

type SqlModelArg struct {
	Log bool `flag:"log"`

	Path string `arg:"path"`
}

func SqlModel(arg *SqlModelArg) error {
	if arg.Log {
		log.Init(true, "")
	}
	gfile, err := golang.ReadFile(arg.Path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(arg.Path)
	for _, inter := range gfile.Interfaces {
		if inter.Tag.Name != "sql" {
			continue
		}
		err = genSqlModel(dir, inter)
		if err != nil {
			return err
		}
	}
	return nil
}

func genSqlModel(dir string, inter *golang.Interface) error {
	var sqlPaths []string
	for _, opt := range inter.Tag.Options {
		if opt.Value == "" {
			continue
		}
		if opt.Key == "file" {
			sqlPaths = append(sqlPaths, opt.Value)
		}
	}
	if len(sqlPaths) != 1 {
		return fmt.Errorf(`sql-model auto generator`+
			` only support one sql path for each `+
			`interface, found %d for "%s"`, len(sqlPaths),
			inter.Name)
	}
	sqlPath := sqlPaths[0]
	sqlPath = filepath.Join(dir, sqlPath)
	var genFunc func(string, *golang.Interface) error
	_, err := os.Stat(sqlPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		genFunc = newSqlModel
	} else {
		genFunc = updateSqlModel
	}
	start := time.Now()
	err = genFunc(sqlPath, inter)
	if err != nil {
		return err
	}
	log.Infof(`[gen] [sql-model] [%v] %s , interface=%s`,
		time.Since(start), sqlPath, inter.Name)
	return nil
}

func newSqlModel(path string, inter *golang.Interface) error {
	c := new(coder.Coder)
	c.P(0, "-- +gen:sql v=", version.Short)
	c.Empty()
	c.Empty()

	for idx, m := range inter.Methods {
		c.P(0, "-- +gen:method ", m.Name)
		c.P(0, "-- TODO: write SQL here.")
		c.P(0, "-- +gen:end")
		if idx != len(inter.Methods)-1 {
			c.Empty()
			c.Empty()
		}
	}
	return c.WriteFile(path)
}

func updateSqlModel(path string, inter *golang.Interface) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")

	feedBack := func(start int) []string {
		var subLines []string
		idx := start
		for ; idx >= 0; idx-- {
			line := lines[idx]
			if !strings.HasPrefix(line, "--") {
				break
			}
			if line == "-- +gen:end" {
				break
			}
			subLines = append(subLines, line)
		}
		for i, j := 0, len(subLines)-1; i < j; i, j = i+1, j-1 {
			subLines[i], subLines[j] = subLines[j], subLines[i]
		}
		return subLines
	}

	walkForward := func(start int) []string {
		var subLines []string
		idx := start
		for ; idx < len(lines); idx++ {
			line := lines[idx]
			subLines = append(subLines, line)
			if line == "-- +gen:end" {
				break
			}
		}
		return subLines
	}

	sqlFile, err := sql.ReadLines(path, lines)
	if err != nil {
		return err
	}
	oriMap := make(map[string][]string)
	for _, m := range sqlFile.Methods {
		idx := m.LineIdx()
		subLines := feedBack(idx - 1)
		subLines = append(subLines, walkForward(idx)...)
		oriMap[m.Name] = subLines
	}

	c := new(coder.Coder)
	c.P(0, "-- +gen:sql v=", version.Short)
	c.Empty()
	c.Empty()
	for idx, m := range inter.Methods {
		name := m.Name
		oriLines := oriMap[name]
		if len(oriLines) == 0 {
			c.P(0, "-- +gen:method ", name)
			c.P(0, "-- TODO: write SQL here.")
			c.P(0, "-- +gen:end")
		} else {
			for _, oriLine := range oriLines {
				c.P(0, oriLine)
			}
		}
		if idx != len(inter.Methods)-1 {
			c.Empty()
			c.Empty()
		}
	}
	return c.WriteFile(path)
}
