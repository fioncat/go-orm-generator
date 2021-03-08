package cmd

import (
	"fmt"
	"os"

	"github.com/fioncat/go-gendb/cmd/c/cgo"
	"github.com/fioncat/go-gendb/cmd/c/csql"
	"github.com/fioncat/go-gendb/cmd/conn"
	"github.com/fioncat/go-gendb/cmd/gen"
	"github.com/fioncat/go-gendb/misc/cmdt"
	"github.com/fioncat/go-gendb/version"
)

var cmds = make(map[string]*cmdt.Command)

func init() {
	cmds["cgo"] = cgo.Cmder
	cmds["csql"] = csql.Cmder
	cmds["gen"] = gen.Cmder
	cmds["conn"] = conn.Cmder
}

func getCmd(name string) *cmdt.Command {
	cmd := cmds[name]
	if cmd == nil {
		fmt.Printf("unknown command \"%s\"\n", name)
		os.Exit(1)
	}
	return cmd
}

// Execute selects the command that needs to be executed
// and give it the right to execute the program.
func Execute(args []string) {
	args = args[1:]

	if len(args) == 0 {
		showHelp()
		os.Exit(0)
	}

	master := args[0]
	switch master {
	case "-h":
		showHelp()
		os.Exit(0)

	case "help":
		if len(args) == 1 {
			showHelp()
			os.Exit(0)
		}

		cmdName := args[1]
		cmd := getCmd(cmdName)
		cmd.ShowHelp()
		os.Exit(0)

	case "-v", "version":
		fmt.Printf("go-gendb %s\n", version.Full)
		os.Exit(0)
	}

	cmd := getCmd(master)
	args = args[1:]
	if len(args) > 0 && (args[0] == "-h" || args[0] == "help") {
		cmd.ShowHelp()
		os.Exit(0)
	}

	cmd.Execute(args)
}

func showHelp() {
	fmt.Println("help")
}
