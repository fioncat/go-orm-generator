package dbcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/fioncat/go-gendb/dbaccess"
	"github.com/fioncat/go-gendb/dbaccess/dbtypes"
	"github.com/fioncat/go-gendb/misc/cct"
	"github.com/fioncat/go-gendb/misc/col"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/misc/term"
	"github.com/fioncat/go-gendb/misc/trace"
	"github.com/fioncat/go-gendb/scanner"
)

var sliceMu sync.Mutex

type Arg struct {
	Mode     string `flag:"mode"`
	Batch    bool   `flag:"batch"`
	Params   string `flag:"p"`
	Log      bool   `flag:"log"`
	SQLNames string `flag:"sql"`

	ConnKey string `arg:"conn-key"`
	Path    string `arg:"path"`
}

type checkTask struct {
	path     string
	name     string
	sql      string
	prepares []interface{}
}

type checkResult struct {
	path string
	name string
	res  *dbtypes.CheckResult
}

func buildTask(path string, params params, filterSet col.Set, hasFilter bool) ([]*checkTask, error) {
	r, err := scanner.SQLFile(path, false)
	if err != nil {
		return nil, errors.Trace(path, err)
	}
	var sqls []scanner.SQL
	if !hasFilter {
		for _, sql := range r.Sqls {
			if len(sql.Tokens) == 0 {
				continue
			}
			if sql.Tokens[0].Flag == "SELECT" {
				sqls = append(sqls, sql)
			}
		}
	} else {
		for _, sql := range r.Sqls {
			if filterSet.Exists(sql.Name) {
				sqls = append(sqls, sql)
			}
		}
	}
	tasks := make([]*checkTask, len(sqls))
	for i, sql := range sqls {
		phs, err := scanner.SQLPlaceholders(sql.SQL)
		if err != nil {
			return nil, errors.Fmt("parse placeholders "+
				"for '%s' failed: %v", sql.Name, err)
		}

		var reps []interface{}
		for _, repName := range phs.Replaces {
			rep := params[repName]
			if rep == nil {
				log.Infof(`WARN: replace param "%s"`+
					` is empty for sql "%s"`, repName, sql.Name)
				rep = ""
			}
			reps = append(reps, rep)
		}
		sqlString := fmt.Sprintf(phs.SQL, reps...)

		var pres []interface{}
		for _, preName := range phs.Prepares {
			pre := params[preName]
			if pre == nil {
				log.Infof(`WARN: prepare param "%s"`+
					` is empty for sql "%s"`, preName, sql.Name)
				pre = ""
			}
			pres = append(pres, pre)
		}

		tasks[i] = &checkTask{
			path:     sql.Path,
			name:     sql.Name,
			sql:      sqlString,
			prepares: pres,
		}
	}

	return tasks, nil
}

func Run(arg *Arg) bool {
	if arg.Log {
		log.Init(true, "")
	}
	err := dbaccess.SetConn(arg.ConnKey)
	if err != nil {
		errMsg("set connection failed", err)
		return false
	}
	if arg.Mode == "" {
		arg.Mode = "mysql"
	}
	tt := trace.NewTimer("check-sql")
	defer tt.Trace()

	tt.Start("fetch")
	paths, err := getPaths(arg)
	if err != nil {
		errMsg("fetch failed", err)
		return false
	}
	if len(paths) == 0 {
		fmt.Printf("no file fetched, nothing to do.")
		return false
	}

	tt.Start("parse-param")
	params, err := parseParam(arg.Params)
	if err != nil {
		errMsg("parse params failed", err)
		return false
	}
	filterSet, hasFilter := parseSqlNames(arg.SQLNames)

	nWorkers := runtime.NumCPU()
	log.Infof("found %d CPU, will start %d workers",
		nWorkers, nWorkers)

	tt.Start("scan")
	var tasks []*checkTask
	wp := cct.NewPool(len(paths), nWorkers, func(task interface{}) error {
		path := task.(string)
		subTasks, err := buildTask(path, params, filterSet, hasFilter)
		if err != nil {
			return err
		}
		sliceMu.Lock()
		tasks = append(tasks, subTasks...)
		sliceMu.Unlock()
		return nil
	})
	wp.Start()

	for _, path := range paths {
		wp.Do(path)
	}

	if err := wp.Wait(); err != nil {
		errMsg("scan failed", err)
		return false
	}

	if len(tasks) == 0 {
		fmt.Println("no task scanned, nothing to do.")
		return false
	}

	tt.Start("check")
	var results []*checkResult
	wp = cct.NewPool(len(tasks), nWorkers, func(task interface{}) error {
		chkTask := task.(*checkTask)
		cr, err := dbaccess.Check(arg.Mode, chkTask.sql, chkTask.prepares)
		if err != nil {
			return err
		}
		chkResult := &checkResult{
			path: chkTask.path,
			name: chkTask.name,
			res:  cr,
		}
		sliceMu.Lock()
		results = append(results, chkResult)
		sliceMu.Unlock()
		return nil
	})
	wp.Start()
	for _, task := range tasks {
		wp.Do(task)
	}

	if err := wp.Wait(); err != nil {
		errMsg("check sql", err)
		return false
	}

	tt.Start("group-show")
	groups := make(map[string][]*checkResult)
	for _, res := range results {
		if _, ok := groups[res.path]; ok {
			groups[res.path] = append(groups[res.path], res)
			continue
		}
		groups[res.path] = []*checkResult{res}
	}
	groupSlice := make([][]*checkResult, 0, len(groups))
	for _, group := range groups {
		groupSlice = append(groupSlice, group)
	}
	sort.Slice(groupSlice, func(i, j int) bool {
		return groupSlice[i][0].path < groupSlice[j][0].path
	})

	for _, results := range groupSlice {
		sort.Slice(results, func(i, j int) bool {
			return results[i].name < results[j].name
		})
		fmt.Printf("file: %s\n", results[0].path)
		for _, result := range results {
			fmt.Printf("  sql: %s", result.name)
			if result.res.Err != nil {
				err := result.res.Err
				fmt.Printf("\n    %s\n", term.Red("error: "+err.Error()))
				continue
			}
			if len(result.res.Warns) > 0 {
				fmt.Println()
				for _, warn := range result.res.Warns {
					fmt.Printf("    %s\n", term.Warn("warn: "+warn))
				}
			} else {
				fmt.Println(term.Info(" [ok]"))
			}
		}
		fmt.Println()
	}

	return true
}

func errMsg(cause string, err error) {
	fmt.Printf("%s %v\n", term.Red(cause), err)
}

func getPaths(arg *Arg) ([]string, error) {
	if !arg.Batch {
		return []string{arg.Path}, nil
	}
	var paths []string
	err := filepath.Walk(arg.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := filepath.Base(path)
		if strings.HasPrefix(filename, ".") {
			return nil
		}

		if strings.HasSuffix(filename, ".sql") {
			paths = append(paths, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	log.Infof("fetched %d files to check.", len(paths))
	return paths, nil
}

type params map[string]interface{}

func parseParam(p string) (params, error) {
	if p == "" {
		return make(params), nil
	}
	ps := strings.Split(p, ",")
	m := make(params, len(ps))
	for _, pv := range ps {
		kvs := strings.Split(pv, ":")
		if len(kvs) != 2 {
			return nil, errors.Fmt(`key-value "%s" bad format`, pv)
		}
		key := kvs[0]
		val := kvs[1]
		if _, ok := m[key]; ok {
			return nil, errors.Fmt(`key "%s" is duplicate`, key)
		}
		m[key] = val
	}
	return m, nil
}

func parseSqlNames(p string) (col.Set, bool) {
	if p == "" {
		return nil, false
	}
	ps := strings.Split(p, ",")
	set := col.NewSet(len(ps))
	for _, name := range ps {
		set.Add(name)
	}
	return set, true
}
