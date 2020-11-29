package cmd

import "github.com/fioncat/go-gendb/generator/dogen"

var genOp = &Operation{
	Name: "gen",

	Usage: genUsage,
	Help:  genHelp,

	ParamPtr: (*dogen.Arg)(nil),

	Action: func(ctx *Context) bool {
		arg := ctx.Param().(*dogen.Arg)
		dogen.Prepare(arg)
		return dogen.One(arg.Path, arg.ConfPath)
	},
}

const genUsage = `gen [flags] <file-path>`

const genHelp = `
Gen reads the Go source code file, and generates
the specified code according to the "+gendb" tag. The
generated codes are named in the format "zz_generated.*.go".

This process will perform lexical analysis and syntax
analysis on the Go code (or the SQL statement in the
associated sql file). Under certain circumstances, it
will also read the database to obtain the structure of
the data table (if want to read the database, you
must use "--conn-key").

Different types of gendb have different generation rules.
Before the "package" definition of the Go code, the generation
rule type must be specified, otherwise the lexical analysis
process will report an error.

For the specific usage of tags, please refer to the documentation
on the project's Github homepage.

Command Flags:

    <file-path>
                   The Go source file path.
    --conn-key <key>
                   Connection key. It is configured by
                   "go-gendb conn-set".
    --conf-path <file-path>
                   Some generators will use some additional
                   configuration options (configured in json format),
                   use this command to specify the configuration path.
                   Please refer to the documentation of different
                   generators for specific configurations.
    --log
                   Enable the log output. If not set, this command
                   will output nothing (unless the generationg fails).
    --log-path
                   Log output file. By default, the log will be output
                   to the terminal, if this parameter is added, it will
                   be output to the specified file.
    --cache
                   Enable the database cache, go-gendb will cache
                   the data read from the database on the local disk.
                   In this way, the next time it is executed, the data
                   on the disk can be directly read without having to
                   request the database. The data is generally the
                   structure of the table, which is convenient for
                   generating orm code (so it does not take up a lot
                   of space). They will be cached to "~/.gogendb/cache_*".
                   The default cache validity period is 1 hour.
                   If the table structure of the database is modified,
                   and the local cache has not expired, you can manually
                   clear the cache through the command
                   "go-gendb clean-cache".
                   In large projects, if a large number of database
                   requests are required, adding this parameter can
                   greatly increase the speed of code generation.
    --cache-ttl <time>
                   Set the time-to-live for the cache data.
                   Use "s", "m", "h" to represent seconds, minutes,
                   and hours respectively. For example, "2h30m" means
                   that the cache expiration time is 2 hours and 30
                   minutes.

See also: scan, parse, conn-set, clean-cache`
