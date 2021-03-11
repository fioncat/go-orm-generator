package sql_orm

import (
	"fmt"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/compile/orm"
)

type conf struct {
	runPath string
	runName string
}

func defaultConf() *conf {
	return &conf{
		runPath: "github.com/fioncat/go-gendb/api/sql/run",
		runName: "run",
	}
}

type Linker struct{}

func (*Linker) Do(gfile *golang.File) ([]coder.Target, error) {
	file, err := orm.Parse(gfile)
	if err != nil {
		return nil, err
	}

	c := defaultConf()
	for _, opt := range gfile.Options {
		switch opt.Key {
		case "run_path":
			c.runPath = opt.Value

		case "run_name":
			c.runName = opt.Value
		}
	}

	ts := make([]coder.Target, len(file.Results))
	for idx, r := range file.Results {
		t := new(target)
		t.path = gfile.Path
		t.r = r
		t.c = c
		t.operName = fmt.Sprintf("%sOper", r.Name)
		t.operType = fmt.Sprintf("_%s", t.operName)

		ts[idx] = t
	}

	return ts, nil
}
