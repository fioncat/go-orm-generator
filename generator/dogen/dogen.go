package dogen

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fioncat/go-gendb/dbaccess"
	"github.com/fioncat/go-gendb/generator"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/misc/term"
	"github.com/fioncat/go-gendb/misc/trace"
	"github.com/fioncat/go-gendb/parser/common"
	"github.com/fioncat/go-gendb/parser/gosql"
	"github.com/fioncat/go-gendb/scanner"
	"github.com/fioncat/go-gendb/store"
)

type Arg struct {
	Cache    bool   `flag:"cache"`
	CacheTTL string `flag:"cache-ttl"`
	Log      bool   `flag:"log"`
	LogPath  string `flag:"log-path"`
	ConfPath string `flag:"conf-path"`
	ConnKey  string `flag:"conn-key"`

	Path string `arg:"path"`
}

func Prepare(arg *Arg) {
	if arg.Log {
		log.Init(true, arg.LogPath)
	}
	if arg.Cache {
		store.EnableCache()
	}
	if arg.CacheTTL != "" {
		ttl, err := time.ParseDuration(arg.CacheTTL)
		if err != nil {
			fmt.Printf(`time "%s" bad format`, arg.CacheTTL)
			fmt.Println()
			os.Exit(1)
		}
		store.SetCacheTTL(ttl)
	}

	if arg.ConnKey != "" {
		err := dbaccess.SetConn(arg.ConnKey)
		if err != nil {
			fmt.Printf("get connection %s failed: %v",
				arg.ConnKey, err)
			os.Exit(1)
		}
	}
}

func One(path, confPath string) bool {
	tt := trace.NewTimer("gen:" + path)
	defer tt.Trace()

	tt.Start("scan")
	sr, err := scanner.Go(path, false)
	if err != nil {
		return onErr(err)
	}

	tt.Start("parse")
	var results []common.Result
	switch sr.Type {
	case "sql":
		results, err = gosql.Parse(sr, false)
	default:
		fmt.Printf("Unknown gen type %s", sr.Type)
		return false
	}

	tt.Start("generate")
	dir := filepath.Dir(path)
	err = generator.Do(path, dir, sr.Package, confPath, results)
	if err != nil {
		return onErr(err)
	}

	return true
}

func onErr(err error) bool {
	fmt.Printf("%s %v\n", term.Red("[error]"), err)
	return false
}
