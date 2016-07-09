package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:  "plan",
			Usage: "Plan DAB whaleprint",
			ArgsUsage: `STACK

Prints an execultion plan to review before applying changes.
Whaleprint will use the stack name to load the DAB file.
			`,
			Action: plan,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Usage: "DAB file to use",
				},
			},
		},
		{
			Name:  "apply",
			Usage: "Apply DAB whaleprint",
			ArgsUsage: `STACK

Prints an execultion plan to review before applying changes.
Whaleprint will use the stack name to load the DAB file.
			`,
			Action: apply,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Usage: "DAB file to use",
				},
			},
		},
	}

	app.Run(os.Args)
}
