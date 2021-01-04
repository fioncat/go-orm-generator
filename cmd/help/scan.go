package help

const ScanUsage = `scan <mode> <file-path>`

const Scan = `
Scan performs lexical analysis on the target file
and outputs the result in the form of json.

Lexical analysis is the first step in generating
code. Taking Go code as an example, Scan will
analyze various "tags" in it and generate a general
analysis structure. Scan only scans, and does not
check its grammar too much.

This command is mainly used for debugging. If the
code generation fails, and the reason is not clear,
you can use this command to view the results of
lexical analysis to troubleshoot the error.

Perform lexical analysis on a Go code file that
has been tagged with +gendb, and execute:
go-gendb scan go /path/to/scan.go

This will output json-format data to the terminal.
Now supported modes:

  go    golang code tagged with "+gendb {args}"
  sql   sql file tagged with "-- !{name}"

See also: parse, gen`
