package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/cli/compose/loader"
	composetypes "github.com/docker/docker/cli/compose/types"
	"github.com/pkg/errors"
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

	stackFiles := []string{}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yml") {
			stackFiles = append(stackFiles, strings.TrimSuffix(file.Name(), ".yml"))
		}
	}

	if len(stackFiles) == 0 {
		log.Fatal("No stacks found in current directory")
	}

	return stackFiles
}

func getStacks(c *cli.Context) ([]Stack, error) {
	type stackDefinition struct {
		name string
		file string
	}

	defs := []stackDefinition{}

	stackNames := c.Args()
	stackFile := c.String("file")

	if stackFile != "" {
		if len(stackNames) > 1 {
			return nil, cli.NewExitError("You can only specify one stack name when using -f", 1)
		} else if len(stackNames) == 1 {
			defs = append(defs, stackDefinition{name: stackNames[0], file: stackFile})
		} else {
			stackName := strings.TrimSuffix(filepath.Base(stackFile), filepath.Ext(stackFile))
			defs = append(defs, stackDefinition{name: stackName, file: stackFile})
		}
	} else if len(stackNames) == 0 {
		stackNames = getStacksFromCWD()

		for _, name := range stackNames {
			stackFile := fmt.Sprintf("%s.yml", name)
			defs = append(defs, stackDefinition{name: name, file: stackFile})
		}
	} else if len(stackNames) > 0 {
		for _, name := range stackNames {
			stackFile := fmt.Sprintf("%s.yml", name)
			defs = append(defs, stackDefinition{name: name, file: stackFile})
		}
	}

	stacks := make([]Stack, len(defs))
	for i, def := range defs {

		configFile, err := getConfigFile(def.file)
		if err != nil {
			return nil, cli.NewExitError(err.Error(), 3)
		}

		details, err := getConfigDetails(configFile)
		if err != nil {
			return nil, cli.NewExitError(err.Error(), 3)
		}

		config, err := loader.Load(details)

		if err != nil {
			return nil, cli.NewExitError(err.Error(), 3)
		}
		stacks[i] = Stack{Name: def.name, Config: config}
	}
	return stacks, nil
}

func getConfigDetails(file *composetypes.ConfigFile) (composetypes.ConfigDetails, error) {
	var details composetypes.ConfigDetails
	var err error

	details.WorkingDir, err = os.Getwd()
	if err != nil {
		return details, err
	}

	details.ConfigFiles = []composetypes.ConfigFile{*file}
	details.Environment, err = buildEnvironment(os.Environ())
	if err != nil {
		return details, err
	}
	return details, nil
}
func buildEnvironment(env []string) (map[string]string, error) {
	result := make(map[string]string, len(env))
	for _, s := range env {
		// if value is empty, s is like "K=", not "K".
		if !strings.Contains(s, "=") {
			return result, errors.Errorf("unexpected environment %q", s)
		}
		kv := strings.SplitN(s, "=", 2)
		result[kv[0]] = kv[1]
	}
	return result, nil
}

func getConfigFile(filename string) (*composetypes.ConfigFile, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	config, err := loader.ParseYAML(bytes)
	if err != nil {
		return nil, err
	}
	return &composetypes.ConfigFile{
		Filename: filename,
		Config:   config,
	}, nil
}
