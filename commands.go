package main

import (
	"fmt"
	"os"

	"github.com/Monitob/edl/command"
	"github.com/codegangsta/cli"
)

var GlobalFlags = []cli.Flag{}

var Commands = []cli.Command{
	{
		Name:   "conf",
		Usage:  "<name>.edl --folder name",
		Action: command.CmdConf,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "folder",
				Value: "Conformation",
				Usage: "Name of the folder to edl result",
			},
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
