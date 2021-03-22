package orm_sql

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/compile/orm"
	"github.com/fioncat/go-gendb/misc/log"
)

const (
	runPath = "run_path"
	runName = "run_name"
	dbUse   = "db_use"
	sqlPath = "sql_path"
)

type Linker struct{}

func (*Linker) DefaultConf() map[string]string {
	return map[string]string{
		runPath: "github.com/fioncat/go-gendb/api/sql/run",
		runName: "run",
		dbUse:   "db",
		sqlPath: "",
		"db":    "",
	}
}

func (*Linker) Do(gfile *golang.File, conf map[string]string) (
	[]coder.Target, error,
) {
	start := time.Now()
	rs, err := orm.Parse(gfile, false)
	if err != nil {
		return nil, err
	}

	if conf[sqlPath] != "" {
		dir := filepath.Dir(gfile.Path)
		path := filepath.Join(dir, conf[sqlPath])
		err = writeCreateTables(conf["db"],
			path, rs)
		if err != nil {
			return nil, err
		}
		log.Infof("[linker] [sql-orm] write create sql to %s", path)
	}

	ts := make([]coder.Target, len(rs))
	for idx, r := range rs {
		t := new(target)
		t.path = gfile.Path
		t.r = r
		t.conf = conf
		t.operName = fmt.Sprintf("%sOper", r.Name)
		t.operType = fmt.Sprintf("_%s", t.operName)

		ts[idx] = t
	}
	log.Infof("[linker] [orm-sql] [%v] %s, %d target(s)",
		time.Since(start), gfile.Path, len(ts))

	return ts, nil
}
