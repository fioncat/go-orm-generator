package help

import "fmt"

func Show() {
	fmt.Println(help)
}

const help = `go-gendb is a command-tool to generate GO database code.

Usage:
    go-gendb <command> [flags]

Generate Commands:
    gen          Generate code for one file.
    batch        Scan directory and generate multi-files.
    clean        Remove generated code(s).
    clean-cache  Clean cached data.
    cache-size   Show the disk space occupied by the cache.

Connection Commands:
    conn-set     Set database connection.
    conn-get     Show database connection.
    conn-del     Remove database connection.

Database Tools:
    check    Check the errors or warnings for sql.
    exec     Execute sql statement.

Debug Commands:
    scan    Perform lexical analysis on the file.
    parse   Perform syntax analysis on the code.

Use "go-gendb help <command>" for more information about a command.`
