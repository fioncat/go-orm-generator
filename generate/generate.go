package generate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/fioncat/go-gendb/build"
	"github.com/fioncat/go-gendb/compile/mediate"
	"github.com/fioncat/go-gendb/compile/parse"
	"github.com/fioncat/go-gendb/compile/scan/scango"
	"github.com/fioncat/go-gendb/compile/token"
	"github.com/fioncat/go-gendb/database/rdb"
	"github.com/fioncat/go-gendb/generate/coder"
	"github.com/fioncat/go-gendb/generate/internal/gensql"
	"github.com/fioncat/go-gendb/generate/internal/genstruct"
	"github.com/fioncat/go-gendb/misc/errors"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/misc/trace"
	"github.com/fioncat/go-gendb/misc/workerpool"
)

type generator interface {
	Name() string
	ConfType() interface{}
	DefaultConf() interface{}
	Do(c *coder.Coder, result mediate.Result, confv interface{}) error
}

type Arg struct {
	Cache    bool   `flag:"cache"`
	CacheTTL string `flag:"cache-ttl"`
	Log      bool   `flag:"log"`
	LogPath  string `flag:"log-path"`
	ConfPath string `flag:"conf-path"`

	Conn   string `flag:"conn"`
	DbType string `flag:"db-type" default:"mysql"`

	Path string `arg:"path"`
}

func prepare(arg *Arg) error {
	if arg.Log {
		log.Init(true, arg.LogPath)
	}
	if arg.Cache {
		rdb.ENABLE_TABLE_CACHE = true
	}
	if arg.CacheTTL != "" {
		// parse and set table cache TimeToLive
		ttl, err := time.ParseDuration(arg.CacheTTL)
		if err != nil {
			return fmt.Errorf("ttl bad format: %s", arg.CacheTTL)
		}
		rdb.TABLE_CACHE_TTL = ttl
	}

	if arg.Conn != "" {
		// need to connect database
		err := rdb.Init(arg.Conn, arg.DbType)
		if err != nil {
			return errors.Trace("init database", err)
		}
	}
	return nil
}

func One(arg *Arg) error {
	err := prepare(arg)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(arg.Path)
	if err != nil {
		return err
	}
	var confData []byte
	if arg.ConfPath != "" {
		confData, err = ioutil.ReadFile(arg.ConfPath)
		if err != nil {
			return err
		}
	}
	return one(arg.Path, data, confData)
}

func one(path string, data, confData []byte) error {
	tt := trace.NewTimer("gen:" + path)
	defer tt.Trace()

	tt.Start("scan")
	sr, err := scango.Do(path, string(data))
	if err != nil {
		return err
	}

	tt.Start("parse")
	results, err := parse.Do(sr)
	if err != nil {
		return err
	}

	tt.Start("generate")
	dir := filepath.Dir(path)
	return do(path, dir, sr.Package, confData, results)
}

func Batch(arg *Arg) error {
	err := prepare(arg)
	if err != nil {
		return err
	}
	var confData []byte
	if arg.ConfPath != "" {
		confData, err = ioutil.ReadFile(arg.ConfPath)
		if err != nil {
			return err
		}
	}
	return batch(arg.Path, confData)
}

func batch(root string, confData []byte) error {
	var paths []string
	tt := trace.NewTimer("batch:" + root)
	defer tt.Trace()

	tt.Start("fetch")
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := filepath.Base(path)
		if strings.HasSuffix(filename, ".go") {
			paths = append(paths, path)
		}

		return nil
	})

	if err != nil {
		return errors.Trace("fetch files", err)
	}

	if len(paths) == 0 {
		return fmt.Errorf("no file to generate")
	}

	tt.Start("generate")
	wp := workerpool.New(len(paths),
		build.N_WORKERS, genWorker(confData))
	wp.Start()
	for _, path := range paths {
		wp.Do(path)
	}

	if err := wp.Wait(); err != nil {
		return err
	}

	return nil
}

func genWorker(confData []byte) workerpool.WorkFunc {
	return func(task interface{}) error {
		path := task.(string)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		lines := strings.Split(content, "\n")
		hasTag := false
		for _, line := range lines {
			if token.TAG_PREFIX.Prefix(line) {
				hasTag = true
				break
			}
			if token.GO_PACKAGE.Prefix(line) {
				break
			}
		}
		if !hasTag {
			// no accept
			return nil
		}

		return one(path, data, confData)
	}
}

var (
	generators   = make(map[string]generator)
	genConfTypes = make(map[string]reflect.Type)

	genConfs sync.Map
)

func init() {
	generators["db-oper"] = &gensql.Generator{}
	generators["struct"] = &genstruct.Generator{}

	for key, gen := range generators {
		t := gen.ConfType()
		if t == nil {
			continue
		}
		genConfTypes[key] = reflect.TypeOf(gen.ConfType()).Elem()
	}
}

func do(path, dir, pkg string, confData []byte, results []mediate.Result) error {
	paths := make([]string, 0, len(results))
	for _, result := range results {
		generator := generators[result.Type()]
		if generator == nil {
			onRemove(paths)
			return fmt.Errorf("can not find "+
				"generator: %s", result.Type())
		}

		c := new(coder.Coder)
		c.Source = path
		c.Pkg = pkg

		confv, _ := genConfs.Load(result.Type())
		if confv == nil && len(confData) > 0 {
			conft, ok := genConfTypes[result.Type()]
			if ok {
				confv = reflect.New(conft).Interface()
				err := json.Unmarshal(confData, confv)
				if err != nil {
					onRemove(paths)
					return errors.Trace("unmarshal conf data", err)
				}
				genConfs.Store(result.Type(), confv)
			}
		}

		if confv == nil {
			confv = generator.DefaultConf()
		}

		err := generator.Do(c, result, confv)
		if err != nil {
			onRemove(paths)
			return errors.Trace("generate", err)
		}

		tarName := fmt.Sprintf("zz_generated_%s.go", result.Key())
		tarPath := filepath.Join(dir, tarName)
		err = c.Write(tarPath)
		if err != nil {
			onRemove(paths)
			return errors.Trace("write target file", err)
		}
		paths = append(paths, tarPath)
	}
	return nil
}

func onRemove(paths []string) {
	for _, path := range paths {
		os.Remove(path)
	}
}
