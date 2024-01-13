package main

import (
	"flag"

	"euphoria.leet.nu/heim/heimctl/cmd"
)

var version string

func main() {
	if version != "" {
		cmd.Version = version
	}
	flag.Parse()
	cmd.Run(flag.Args())
}
