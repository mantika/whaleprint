package main

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/net/context"

	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

func destroy(c *cli.Context) error {
	stacks, err := getStacks(c)
	if err != nil {
		return err
	}

	var services []string
	for _, stack := range stacks {
		for name, _ := range stack.Bundle.Services {
			services = append(services, fmt.Sprintf("%s_%s", stack.Name, name))
		}
	}

	force := c.Bool("force")

	swarm, swarmErr := client.NewEnvClient()
	if swarmErr != nil {
		return cli.NewExitError(swarmErr.Error(), 3)
	}

	if !force {
		fmt.Printf("Are you sure you want to remove the following services? (%s) yes/no: ", strings.Join(services, ", "))
		var input string
		fmt.Scanln(&input)
		switch {
		case input == "no":
			log.Fatal("Aborting")
		case input == "yes":
		default:
			log.Fatalf("Incorrect option \"%s\", aborting", input)
		}
	}

	for _, service := range services {
		color.Cyan("Removing service %s\n", service)
		servicesErr := swarm.ServiceRemove(context.Background(), service)
		if servicesErr != nil {
			log.Println("Error removing service %s:, %s", service, err)
		}
	}

	return nil
}
