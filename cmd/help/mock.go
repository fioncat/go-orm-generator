package help

const MockUsage = `mock [flags] <path>`

const Mock = `
Mock is used to parse mock-toml files and mock sql statements.

The mock-toml file is a simple toml configuration file. It can
contain some special mock rules. Using it, you can easily mock
out the data in the format you want. For details about mock-toml,
please refer to the online documentation.

You can choose to store the mock data locally(in sql file) and put it
in the database for execution manually, or you can execute the mock
command directly in the database. We recommend the former more so
that you can check the sql statement before inserting the mock data.

If neither "--exec" nor "--file" is specified, the mock data will be
printed on the terminal.

Command Flags:

    <path>
             The path of the mock-toml file.
    --exec
             Execute sql statements directly in the configured database
             to insert mock data. This requires you to configure "conn"
             in mock-toml.
    --mute
             Do not output any process information in the terminal
             during the mock process.

    --file <file-path>
             Output the mock data in the specified file in the form of
             sql statement. We recommend using this parameter to generate
             a sql file and manually execute it in the database.

See alse: exec, check, conn-set`
