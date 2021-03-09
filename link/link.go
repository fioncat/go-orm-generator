package link

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/link/internal/sql"
	"github.com/fioncat/go-gendb/misc/errors"
)

type linker interface {
	Do(file *golang.File) ([]coder.Target, error)
}

type Result struct {
	Targets []coder.Target
	Imports []golang.Import

	Package string
}

var linkers map[string]linker

func init() {
	linkers = map[string]linker{
		"sql": &sql.Linker{},
	}
}

func Do(file *golang.File) (*Result, error) {
	linker := linkers[file.Type]
	if linker == nil {
		return nil, fmt.Errorf(`can not find `+
			`linker "%s"`, file.Type)
	}
	res := new(Result)
	res.Package = file.Package
	for _, opt := range file.Options {
		switch opt.Key {
		case "import":
			var imp golang.Import
			imp.Path = opt.Value
			res.Imports = append(res.Imports, imp)

		case "conn":
			var connName string
			var connType = "mysql"
			tmp := strings.Split(opt.Value, ",")
			switch len(tmp) {
			case 1:
				connName = opt.Value

			case 2:
				connType = tmp[0]
				connName = tmp[1]

			default:
				return nil, fmt.Errorf(`conn `+
					`config "%s" is bad format`, opt.Value)
			}
			err := rdb.Init(connName, connType)
			if err != nil {
				return nil, errors.Trace("connect database", err)
			}

		case "package":
			res.Package = opt.Value
		}
	}

	ts, err := linker.Do(file)
	if err != nil {
		return nil, err
	}
	res.Targets = ts

	return res, nil
}
