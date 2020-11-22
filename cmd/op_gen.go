package cmd

import (
	"fmt"
	"time"

	"github.com/fioncat/go-gendb/dbaccess"
	"github.com/fioncat/go-gendb/generator/genfile"
	"github.com/fioncat/go-gendb/generator/gensql"
	"github.com/fioncat/go-gendb/misc/log"
	"github.com/fioncat/go-gendb/misc/term"
	"github.com/fioncat/go-gendb/store"
)

type genArg struct {
	LogPath     string `flag:"log-file"`
	EnableLog   bool   `flag:"log"`
	EnableCache bool   `flag:"cache"`
	CacheTTL    string `flag:"cache-ttl"`
	ConnKey     string `flag:"conn"`
	Runner      string `flag:"runner"`

	Path string `arg:"path"`
}

var genOp = &Operation{
	Name:     "gen",
	ParamPtr: (*genArg)(nil),

	Usage: genUsage,
	Help:  genHelp,

	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*genArg)
		if arg.EnableLog {
			log.Init(true, arg.LogPath)
		}
		if arg.EnableCache {
			store.EnableCache()
		}
		if arg.Runner != "" {
			gensql.SetRunnerPath(arg.Runner)
		}

		if arg.CacheTTL != "" {
			ttl, err := time.ParseDuration(arg.CacheTTL)
			if err != nil {
				fmt.Printf(`time "%s" bad format`, arg.CacheTTL)
				fmt.Println()
				return false
			}
			store.SetCacheTTL(ttl)
		}
		if arg.ConnKey != "" {
			err := dbaccess.SetConn(arg.ConnKey)
			if err != nil {
				fmt.Printf("get connection %s failed: %v",
					arg.ConnKey, err)
				return false
			}
		}
		err := genfile.Do(arg.Path)
		if err != nil {
			fmt.Printf("%s %v\n", term.Red("[error]"), err)
			return false
		}
		return true
	},
}

const genUsage = `gen [flags] <file-path>`

const genHelp = `
Read the go code file, parse the code to generate tags,
and generate code in the same level directory of the
input file.

Before entering the package definition of the code file,
there must be a "// +gendb {type}" tag. Otherwise, the
file will be rejected.

In the execution process, the program may need to read
the table structure of the remote database. This requires
setting up the database connection through the
"go-gendb conn-set" command in advance, and importing it
through "--conn" when using this command. Otherwise,
"connection is not set" error will be reported.

Gen can cache the table structure locally, so that in
the subsequent generation process, the table structure
can be read directly from the disk without reading the
database. This behavior is enabled by "--cache". The
default cache time is 1 hour, and the cache time can
be modified by "--cache-ttl". If the database table
changes and the cache has not expired, you can also
manually clear the cache through "go-gendb clean-cache".

For detailed rules about gendb tags, please refer to the
online documentation.

Command Flags:

    <file-path>
            input file path.
    --log
            show logs when generating.
    --log-file
            write log to a file.
    --cache
            Try to read table structure from the
            disk. If cache miss, will still read it from
            the database, if success, the table structure
            will be save to the disk. So the next time
            gen no need to interact with the database,
            instead directly read them from the disk.
            If the project is very large, add this flag
            can greatly improve the code generation time.
            The cache data will be saved to "~/.gogendb/cache_{key}"
    --cache-ttl <time>
            Set the expiration time of the cache, the
            default is 1 hour.
    --conn-key <key>
            If the database need to be access during the
            code generation process, must specify the database
            connection through this flag. The database connection
            is configured through "go-gendb set-conn".
    --runner <runner-path>
            The generated code needs to use the runner
            package. This parameter specifies the go import path
            of the runner package.
            The default is "github.com/fioncat/go-gendb/api/sqlrunner".
            If you don't want to introduce go-gendb into the
            project, you can copy the code in sqlrunner directly
            to the project and import it through this parameter.

See https://github.com/fioncat/go-gendb for more information.`
