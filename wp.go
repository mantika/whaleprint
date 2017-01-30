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
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "Manage DAB files as docker swarm service blueprints"
	app.Version = "0.0.3"

	app.Commands = []cli.Command{
		{
			Name:  "plan",
			Usage: "Plan service deployment",
			ArgsUsage: `[STACK] [STACK...]

Prints an execultion plan to review before applying changes.
Whaleprint will look for .dab files or use the stack name to load the DAB file.
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
			Usage: "Apply service deployment",
			ArgsUsage: `[STACK] [STACK...]

Applies the execution plan returned by the "whaleprint plan" command
Whaleprint will look for .dab files or use the stack name to load the DAB file.
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
				cli.StringSliceFlag{
					Name:  "with-registry-auth",
					Usage: "Registry auth parameter for private images",
				},
			},
		},
		{
			Name:  "export",
			Usage: "Export stacks to DAB",
			ArgsUsage: `
Exports current service definitions to a DAB file
			`,
			Action: export,
		},
		{
			Name:  "destroy",
			Usage: "Destroy a DAB stack",
			ArgsUsage: `[STACK] [STACK...]

Destroys the stack present in the DAB file  
Whaleprint will look for .dab files use the stack name to load the DAB file.
			`,
			Action: destroy,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Usage: "DAB file to use",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "Ignore destroy DAB file to useconfirmation",
				},
				cli.StringSliceFlag{
					Name:  "target",
					Usage: "Process specified services only (default [])",
				},
			},
		},
		{
			Name:  "output",
			Usage: "Show import output information stacks",
			ArgsUsage: `[STACK] [STACK...]

Show important information for the specified stacks.
			`,
			Action: output,
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

func getStacksFromCWD() []string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal("Error fetching files from current dir", err)
	}

	dabs := []string{}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".dab") {
			dabs = append(dabs, strings.TrimSuffix(file.Name(), ".dab"))
		}
	}

	if len(dabs) == 0 {
		log.Fatal("No DABs found in current directory")
	}

	return dabs
}

func getStacks(c *cli.Context) ([]Stack, error) {
	type stackDefinition struct {
		name string
		file string
	}

	defs := []stackDefinition{}

	stackNames := c.Args()
	dabFile := c.String("file")

	if dabFile != "" {
		if len(stackNames) > 1 {
			return nil, cli.NewExitError("You can only specify one stack name when using -f", 1)
		} else if len(stackNames) == 1 {
			defs = append(defs, stackDefinition{name: stackNames[0], file: dabFile})
		} else {
			stackName := strings.TrimSuffix(filepath.Base(dabFile), filepath.Ext(dabFile))
			defs = append(defs, stackDefinition{name: stackName, file: dabFile})
		}
	} else if len(stackNames) == 0 {
		stackNames = getStacksFromCWD()

		for _, name := range stackNames {
			dabFile := fmt.Sprintf("%s.dab", name)
			defs = append(defs, stackDefinition{name: name, file: dabFile})
		}
	} else if len(stackNames) > 0 {
		for _, name := range stackNames {
			dabFile := fmt.Sprintf("%s.dab", name)
			defs = append(defs, stackDefinition{name: name, file: dabFile})
		}
	}

	stacks := make([]Stack, len(defs))
	for i, def := range defs {
		var dabReader io.Reader
		if u, e := url.Parse(def.file); e == nil && u.IsAbs() {
			// DAB file seems to be remote, try to download it first
			return nil, cli.NewExitError("Not implemented", 2)
		} else {
			if file, err := os.Open(def.file); err != nil {
				return nil, cli.NewExitError(err.Error(), 3)
			} else {
				dabReader = file
			}
		}

		bundle, bundleErr := bundlefile.LoadFile(dabReader)
		if bundleErr != nil {
			return nil, cli.NewExitError(bundleErr.Error(), 3)
		}
		stacks[i] = Stack{Name: def.name, Bundle: bundle}
	}
	return stacks, nil
}
