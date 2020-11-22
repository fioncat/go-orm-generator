package cmd

const help = `go-gendb is a command-tool to generate GO database code.

Usage:

    go-gendb <command> [flags]

Commands:

    gen          Generate code for one file.
    batch        Scan directory and generate multi-files.
    clean        Remove generated code(s).
    conn-set     Set database connection.
    conn-get     Show database connection.
    conn-del     Remove database connection.
    clean-cache  Clean cached data.

Use "go-gendb help <command>" for more information about a command.`
