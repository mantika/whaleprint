package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/client/bundlefile"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:  "plan",
			Usage: "Plan DAB whaleprint",
			ArgsUsage: `[STACK]

Prints an execultion plan to review before applying changes.
Whaleprint will use the stack name to load the DAB file.
			`,
			Action: plan,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Usage: "DAB file to use",
				},
				cli.BoolFlag{
					Name:  "detail",
					Usage: "Show all properties instead of changes only",
				},
				cli.StringSliceFlag{
					Name:  "target",
					Usage: "Process specified services only (default [])",
				},
			},
		},
		{
			Name:  "apply",
			Usage: "Apply DAB whaleprint",
			ArgsUsage: `[STACK]

Prints an execultion plan to review before applying changes.
Whaleprint will use the stack name to load the DAB file.
			`,
			Action: apply,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Usage: "DAB file to use",
				},
				cli.StringSliceFlag{
					Name:  "target",
					Usage: "Process specified services only (default [])",
				},
			},
		},
	}

	app.Run(os.Args)
}

func getStacFromCWD() string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal("Error fetching files from current dir", err)
	}

	var dab string

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".dab") {
			if dab != "" {
				color.Yellow("%s", "[WARN] multiple DAB files found in CWD, use stackname or -f flag \n")
				os.Exit(1)
			}
			dab = strings.TrimSuffix(file.Name(), ".dab")
		}
	}

	if dab == "" {
		log.Fatal("No DAB found in current directory")
	}

	return dab
}

func getBundleFromContext(c *cli.Context) (*bundlefile.Bundlefile, string, error) {

	stackName := c.Args().Get(0)
	dabFile := c.String("file")

	if stackName == "" && dabFile == "" {
		stackName = getStacFromCWD()
	}

	if dabFile != "" {
		// Assume it is called as the stack name
		stackName = strings.TrimSuffix(filepath.Base(dabFile), filepath.Ext(dabFile))
		fmt.Println(stackName, "lalala")
	} else {
		dabFile = fmt.Sprintf("%s.dab", stackName)
	}

	var dabReader io.Reader
	if u, e := url.Parse(dabFile); e == nil && u.IsAbs() {
		// DAB file seems to be remote, try to download it first
		return nil, "", cli.NewExitError("Not implemented", 2)
	} else {
		if file, err := os.Open(dabFile); err != nil {
			return nil, "", cli.NewExitError(err.Error(), 3)
		} else {
			dabReader = file
		}
	}

	bundle, bundleErr := bundlefile.LoadFile(dabReader)
	if bundleErr != nil {
		return nil, "", cli.NewExitError(bundleErr.Error(), 3)
	}
	return bundle, stackName, nil
}
