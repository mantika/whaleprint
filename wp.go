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
			Usage:  "Plan usage",
			Action: plan,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Usage: "DAB file to use",
				},
			},
		},
		{
			Name:   "apply",
			Usage:  "Apply usage",
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
