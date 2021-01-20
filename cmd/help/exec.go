package help

const ExecUsage = `exec [flags] <conn> <path> <method>`

const Exec = `
exec is used to directly execute sql statements in a certain
sql file, and the sql file must meet the go-gendb format. The
statements to be executed can be query statements or execution
statements, and their execution results will be printed directly
on the terminal. Note that the execution of dynamic statements
is not supported (defined by "-- +", there are "+{xxx}" placeholders).

For the placeholders "${name}" and "#{name}" in the SQL statement,
they are specified in json format through the parameter "m". If not
provided, you will be asked to enter them manually at the terminal
before executing the SQL.

The database connection key must be specified through <conn>. This is
configured through the "conn-set" command.

Command Flags:

     <conn>
                      The connection key configured by "conn-set".
     <path>
                      The path of sql file.
     <method>
                      The method name of the sql statement.
     -m <json-params>
                      Json-format data, use for replace placeholders.
     --fmt <fmt-type>
                      The format type of query output(default "fields"),
                      options:
                        "fields": each row occupies multiple lines.
                        "json":   each row is a json-format data.
                        "table":  each row occupied one line.
     --db-type <db-type>
                      The database type, default "mysql"
     --rows-limit <limit>
                      Max number of rows displayed in the query(default 100)

See also: check, gen, conn-set`
