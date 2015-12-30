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
		Usage:  "edl conf --file-edl <file> --dir <Dir> --project-name<Folder>",
		Action: command.CmdConf,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "e, file-edl",
				Usage: "Specify a file edl",
			},
			cli.StringFlag{
				Name:  "d, dir",
				Usage: "Specify a directory to search(default: root )",
			},
			cli.StringFlag{
				Name:  "p, project-name",
				Usage: "Specify an alternate project name(default: conformation)",
			},
		},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
