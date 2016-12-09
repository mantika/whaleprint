package main

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func output(c *cli.Context) error {
	stacks, err := getStacks(c)
	if err != nil {
		return err
	}

	swarm, swarmErr := client.NewEnvClient()
	if swarmErr != nil {
		return cli.NewExitError(swarmErr.Error(), 3)
	}

	for _, stack := range stacks {
		filter := filters.NewArgs()
		filter.Add("label", "com.docker.stack.namespace="+stack.Name)
		services, servicesErr := swarm.ServiceList(context.Background(), types.ServiceListOptions{Filters: filter})
		if servicesErr != nil {
			return cli.NewExitError(servicesErr.Error(), 3)
		}

		current := getSwarmServicesSpecForStack(services)

		for _, s := range current {
			color.Green("%s\n", s.Spec.Name)
			fmt.Println("  - Published Ports")

			for _, port := range s.Endpoint.Ports {
				fmt.Printf("     %d => %d\n", port.PublishedPort, port.TargetPort)
			}

			fmt.Println()
		}
	}

	return nil
}
