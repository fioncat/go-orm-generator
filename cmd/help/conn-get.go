package help

const ConnGetUsage = `conn-get <key>`

const ConnGet = `
Show the database connection in json-form.

Specify the connection to be showed through
the connection key.

For example, to show the "test" connection can use:

go-gendb conn-get test

See also: conn-set, conn-del`
