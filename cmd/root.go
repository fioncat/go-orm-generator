package cmd

import (
	"fmt"
	"os"

	"github.com/fioncat/go-gendb/build"
	"github.com/fioncat/go-gendb/cmd/help"
	"github.com/fioncat/go-gendb/misc/cmdt"
)

var cmds = make(map[string]*cmdt.Command)

func init() {
	cmds["scan"] = scanCmd
	cmds["parse"] = parseCmd
	cmds["conn-set"] = setConnCmd
	cmds["conn-get"] = getConnCmd
	cmds["gen"] = genCmd
	cmds["batch"] = batchCmd
	cmds["cache-size"] = cacheSizeCmd
	cmds["clean-cache"] = cleanCacheCmd
	cmds["clean"] = cleanCmd
	cmds["check"] = checkCmd
	cmds["exec"] = execCmd
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
		help.Show()
		os.Exit(0)
	}

	master := args[0]
	switch master {
	case "-h":
		help.Show()
		os.Exit(0)

	case "help":
		if len(args) == 1 {
			help.Show()
			os.Exit(0)
		}

		cmdName := args[1]
		cmd := getCmd(cmdName)
		cmd.ShowHelp()
		os.Exit(0)

	case "-v", "version":
		build.ShowVersion()
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
