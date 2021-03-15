package orm_mgo

import (
	"time"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/compile/orm"
	"github.com/fioncat/go-gendb/misc/log"
)

type Linker struct{}

func (*Linker) DefaultConf() map[string]string {
	return map[string]string{}
}

func (*Linker) Do(gfile *golang.File, conf map[string]string) (
	[]coder.Target, error,
) {
	start := time.Now()
	rs, err := orm.Parse(gfile, true)
	if err != nil {
		return nil, err
	}

	ts := make([]coder.Target, len(rs))
	for idx, r := range rs {
		t := new(target)
		t.path = gfile.Path
		t.r = r
		t.conf = conf
		ts[idx] = t
	}
	log.Infof("[linker] [orm-mgo] [%v] %s, %d target(s)",
		time.Since(start), gfile.Path, len(ts))

	return ts, nil
}
