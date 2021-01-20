package help

var CheckUsage = `check [flags] <conn> <path>`

var Check = `
check is used to check the sql statement of the sql file.
It can check out the syntax error or warning information
of the sql statement.

The format of the input sql file must meet the go-gendb sql
file format (see the "gen" command documentation for details),
because the database connection needs to be involved in the
inspection process, so you must use the "conn-set" command to
set the connection configuration in advance.

The checked sql statement itself will not be executed, but only
checked, and the check result will be directly output to the
terminal.

The SQL statement can contain placeholders (except dynamic SQL).
The specific value of the placeholder needs to be specified by
parameters. If the value of a placeholder is not specified, an
empty value will be used by default. Please note that this feature
brings potential problems.

Command Flags:

     <conn>
                      Connection key set by "conn-set".
     <path>
                      The path of the sql file(if batch enable, it is
                      the scaning directory).
     --type <db-type>
                      The database type(default "mysql")
     --batch
                      If this flag is enabled, it means that <path> is
                      a directory. In this case, the check command will
                      scan all SQL query statements in all SQL files under
                      path and execute check on them.
     -p <param-path>
                      A file in json format that stores the value
                      of placeholders in sql statements. Before checking
                      a sql statement, the value will be used to replace
                      the placeholder. For example, if json "{"a": 123}"
                      is provided, the "${a}" placeholder will be replaced
                      with "123" when checking. You need to separately define
                      this json in a file and provide the path to this flag.
      --log
                      Enable log
      --filter
                      If you only want to check the specified sql statement,
                      use this flag to specify the name of the sql statement
                      to be checked, separated by ",". All sql query statements
                      are checked by default.

See also: exec, gen, conn-set`
