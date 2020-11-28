package cmd

const help = `go-gendb is a command-tool to generate GO database code.

Usage:

    go-gendb <command> [flags]

Generate Commands:

    gen          Generate code for one file.
    batch        Scan directory and generate multi-files.
    clean        Remove generated code(s).
    clean-cache  Clean cached data.

Connection Commands:

    conn-set     Set database connection.
    conn-get     Show database connection.
    conn-del     Remove database connection.

Debug Commands:

    scan    Perform lexical analysis on the file.
    pase    Perform syntax analysis on the code.

Use "go-gendb help <command>" for more information about a command.`
