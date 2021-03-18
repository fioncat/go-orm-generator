package link

import (
	"fmt"
	"strings"

	"github.com/fioncat/go-gendb/coder"
	"github.com/fioncat/go-gendb/compile/base"
	"github.com/fioncat/go-gendb/compile/golang"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/link/internal/deepcopy"
	"github.com/fioncat/go-gendb/link/internal/orm_mgo"
	"github.com/fioncat/go-gendb/link/internal/orm_sql"
	"github.com/fioncat/go-gendb/link/internal/sql"
	"github.com/fioncat/go-gendb/misc/errors"
)

type linker interface {
	DefaultConf() map[string]string

	Do(file *golang.File, conf map[string]string) ([]coder.Target, error)
}

type extractLinker interface {
	Do(file *golang.File, s *golang.Struct, opts []base.Option) ([]coder.Target, error)
}

type Result struct {
	Targets []coder.Target
	Imports []golang.Import

	Package string
}

var (
	linkers   map[string]linker
	exLinkers map[string]extractLinker
)

func init() {
	linkers = map[string]linker{
		"sql": &sql.Linker{},

		"orm-sql": &orm_sql.Linker{},
		"orm-mgo": &orm_mgo.Linker{},

		"common": &emptyLinker{},
	}

	exLinkers = map[string]extractLinker{
		"deepcopy": &deepcopy.Linker{},
	}
}

func Do(file *golang.File) (*Result, error) {
	linker := linkers[file.Type]
	if linker == nil {
		return nil, fmt.Errorf(`can not find `+
			`linker "%s"`, file.Type)
	}
	conf := linker.DefaultConf()
	if conf == nil {
		conf = make(map[string]string)
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

		default:
			conf[opt.Key] = opt.Value
		}
	}

	ts, err := linker.Do(file, conf)
	if err != nil {
		return nil, err
	}
	res.Targets = ts

	for _, stc := range file.Structs {
		optsGroup := make(map[string][]base.Option, len(stc.Tags))
		for _, tag := range stc.Tags {
			optsGroup[tag.Name] = append(optsGroup[tag.Name],
				tag.Options...)
		}
		for name, opts := range optsGroup {
			exLinker := exLinkers[name]
			if exLinker == nil {
				continue
			}
			ts, err := exLinker.Do(file, stc, opts)
			if err != nil {
				err = errors.Trace(file.Path, err)
				err = errors.OnCompile(file.Path, file.Lines, err)
				return nil, err
			}
			res.Targets = append(res.Targets, ts...)
		}
	}

	return res, nil
}
