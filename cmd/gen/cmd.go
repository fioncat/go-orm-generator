package gen

import (
	"github.com/fioncat/go-gendb/generate"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var Cmder = &cmdt.Command{
	Name: "gen",
	Pv:   (*generate.Arg)(nil),

	Usage: `gen [flags] <path>`,
	Help:  help,

	Action: func(p interface{}) error {
		return generate.Do(p.(*generate.Arg))
	},
}

const help = `
Gen reads the Go source code file, and generates
the specified code according to the "+gen:" tag. The
generated codes are named in the format "zz_generated_*.go".

This process will perform lexical analysis and syntax
analysis on the Go code (or the SQL statement in the
associated sql file). Under certain circumstances, it
will also read the database to obtain the structure of
the data table.

Different types of gendb have different generation rules.
Before the "package" definition of the Go code, the generation
rule type must be specified, otherwise the lexical analysis
process will report an error.

For the specific usage of tags, please refer to the documentation
on the project's Github homepage.

Command Flags:

    <file-path>
                   The Go source file path.
    -o
                   The output directory of the generated file. The
                   default is the same directory of the input file.
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

See also: clean`
