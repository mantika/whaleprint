package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:   "plan",
			Usage:  "TODO!!!",
			Action: plan,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dab, f",
					Usage: "DAB file to use",
				},
			},
		},
		{
			Name:   "apply",
			Usage:  "TODO!!!",
			Action: apply,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dab, f",
					Usage: "DAB file to use",
				},
			},
		},
	}

	app.Run(os.Args)
}
