package help

const ConnSetUsage = `go-gendb conn-set <key> <addr> [flags]`

const ConnSet = `
Set the database connection.

In some cases, the generated code needs to read
database data, and then the database connection
is used, which is configured through this command.
Use the "--conn" option in "go-gendb gen" to
introduce the configured connection.

A connection requires the following configuration:

<key>(required):  The index key of the connection.
                "gen", "conn-get", "conn-del" use this
                key to find the connection.
<addr>(required): The database address.

Optional Flags:

--user:     The database username.
--pass:     The database password.
--db:       The database name.

See alse: gen, conn-get, conn-del
`
