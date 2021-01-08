package help

const BatchUsage = `batch [flags] <root>`

const Batch = `
Batch scans all Go code files in a certain directory,
and if they have a "+gendb" tag, call the generator to
generate codes for it.

The batch parameters and usage are similar to the "gen"
command, the difference is:

- The "<path>" parameter of batch passes the directory
  path. Not a file path.
- Batch will try to generate code for multiple go files.
  If the generation of a go file fails, the entire process
  will be terminated, but the generated code will not be
  deleted. If you want to delete it, please execute
  "go-gendb clean" manually.
- For multiple go files, the generation process will be executed
  concurrently. The number of concurrent worker depends
  on the number of CPU cores of the current machine. Therefore,
  the order of the generated code is different each time. If
  the order is important, do not use this command, but implement
  it by executing "gen" multiple times.

See also: gen`
