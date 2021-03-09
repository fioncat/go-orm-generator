package check

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fioncat/go-gendb/compile/sql"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/database/tools/common"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/misc/term"
)

type Arg struct {
	Log bool `flag:"log"`
	Dir bool `flag:"d"`

	DbType string `flag:"db-type" default:"mysql"`

	Conn string `arg:"conn"`
	Path string `arg:"path"`
}

type checkItem struct {
	Path string
	Name string

	exec *common.Exec

	result rdb.CheckResult
}

func Do(arg *Arg) error {
	if arg.Log {
		log.Init(true, "")
	}
	err := rdb.Init(arg.Conn, arg.DbType)
	if err != nil {
		return err
	}

	var paths []string
	if !arg.Dir {
		paths = []string{arg.Path}
	} else {
		paths, err = fetchDir(arg.Path)
		if err != nil {
			return err
		}
	}

	var items []*checkItem
	for _, path := range paths {
		subItems, err := getToCheck(path)
		if err != nil {
			return err
		}
		items = append(items, subItems...)
	}
	if len(items) == 0 {
		return fmt.Errorf("no method to check")
	}

	for _, item := range items {
		exec := item.exec
		res, err := rdb.Get().Check(exec.Sql, exec.Vals)
		if err != nil {
			return fmt.Errorf(`Check method "%s" `+
				`failed: %v`, item.Name, err)
		}
		item.result = res
	}

	// Show Result
	showResult(paths, items)
	return nil
}

func fetchDir(dir string) ([]string, error) {
	var paths []string
	wf := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := filepath.Base(path)
		if !strings.HasSuffix(filename, ".sql") {
			return nil
		}
		paths = append(paths, path)
		return nil
	}
	err := filepath.Walk(dir, wf)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func getToCheck(path string) ([]*checkItem, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	file, err := sql.ReadLines(path, lines)
	if err != nil {
		return nil, err
	}

	var items []*checkItem
	for _, m := range file.Methods {
		if !common.FindMethodTag(m, "check") {
			continue
		}
		exec, err := common.Method2Exec(m, "check")
		if err != nil {
			err = errors.OnCompile(path, lines, err)
			return nil, err
		}
		item := new(checkItem)
		item.Path = path
		item.Name = m.Name
		item.exec = exec

		items = append(items, item)
	}

	return items, nil
}

func showResult(paths []string, items []*checkItem) {
	group := make(map[string][]*checkItem)
	for _, item := range items {
		group[item.Path] = append(
			group[item.Path], item)
	}
	var errCnt int
	var warnCnt int
	var okCnt int

	for _, path := range paths {
		items := group[path]
		if items == nil {
			continue
		}
		var nameLen int
		for _, item := range items {
			if len(item.Name) > nameLen {
				nameLen = len(item.Name)
			}
		}
		nameFmt := "%-" + strconv.Itoa(nameLen) + "s"

		fmt.Printf("file: %s\n", path)
		for _, item := range items {
			r := item.result
			name := fmt.Sprintf(nameFmt, item.Name)
			fmt.Printf("  sql: %s", name)
			if err := r.GetErr(); err != nil {
				errCnt += 1
				fmt.Printf("\n    %s\n",
					term.Red("error: "+err.Error()))
				continue
			}
			if warns := r.GetWarns(); len(warns) > 0 {
				warnCnt += 1
				for _, warn := range warns {
					fmt.Printf("\n    %s\n", term.Warn(warn))
				}
			} else {
				okCnt += 1
				fmt.Println(term.Info(" [ok]"))
			}
		}
		fmt.Println()
	}
	fmt.Printf("\nCheck done, %s OK, %s Error, %s Warn\n",
		term.Info(strconv.Itoa(okCnt)),
		term.Red(strconv.Itoa(errCnt)),
		term.Warn(strconv.Itoa(warnCnt)))
}
