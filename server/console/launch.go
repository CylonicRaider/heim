package console

import (
	"os"
	"path/filepath"
)

type launcher struct {
	Console
}

func handleLaunchError(con Console, err error) {
	if err != nil {
		con.Printf("FATAL: %s", err)
	}
}

func LaunchCLI(con Console, cli *CLI, argv []string) {
	var c CLI = *cli
	c.Argv = argv
	err := c.Run(launcher{con})
	handleLaunchError(con, err)
}

func LaunchCommand(con Console, cmd *Command, argv []string) {
	err := cmd.Run(launcher{con}, argv)
	handleLaunchError(con, err)
}

func NormalizeProgName(argv0 string) string {
	return filepath.Base(argv0)
}

func LaunchCLIOS(desc string, cli *CLI) {
	LaunchCommandOS(&Command{"", desc, cli})
}

func LaunchCommandOS(cmd *Command) {
	con := NewStdioConsole()
	defer con.Close()

	if cmd.Name == "" {
		cmd.Name = NormalizeProgName(os.Args[0])
	}

	LaunchCommand(con, cmd, os.Args[1:])
}
