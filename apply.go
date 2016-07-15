package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/docker/docker/api/client/bundlefile"
	"github.com/docker/docker/api/client/stack"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/network"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func apply(c *cli.Context) error {
	stackName := c.Args().Get(0)

	if stackName == "" {
		return cli.NewExitError("Need to specify a stack name", 1)
	}
	dabLocation := c.String("dab")

	if dabLocation == "" {
		// Assume it is called as the stack name
		dabLocation = fmt.Sprintf("%s.dab", stackName)
	}

	var dabReader io.Reader
	if u, e := url.Parse(dabLocation); e == nil && u.IsAbs() {
		// DAB file seems to be remote, try to download it first
		return cli.NewExitError("Not implemented", 2)
	} else {
		if dabFile, err := os.Open(dabLocation); err != nil {
			return cli.NewExitError(err.Error(), 3)
		} else {
			dabReader = dabFile
		}
	}

	bundle, bundleErr := bundlefile.LoadFile(dabReader)
	if bundleErr != nil {
		return cli.NewExitError(bundleErr.Error(), 3)
	}

	swarm, swarmErr := client.NewEnvClient()
	if swarmErr != nil {
		return cli.NewExitError(swarmErr.Error(), 3)
	}

	target := c.StringSlice("target")
	targetMap := map[string]bool{}

	for _, name := range target {
		targetMap[name] = true
	}

	services, servicesErr := swarm.ServiceList(context.Background(), types.ServiceListOptions{})
	if servicesErr != nil {
		return cli.NewExitError(servicesErr.Error(), 3)
	}

	expected := getBundleServicesSpec(bundle, stackName)
	current := getSwarmServicesSpecForStack(services, stackName)

	err := updateNetworks(context.Background(), swarm, getUniqueNetworkNames(bundle.Services), stackName)

	if err != nil {
		log.Fatal("Error updating networks when creating services", err)
	}

	cyan := color.New(color.FgCyan)
	for name, expectedService := range expected {
		if _, found := targetMap[expectedService.Spec.Name]; len(targetMap) == 0 || found {
			if currentService, found := current[name]; found {
				// service exists, need to update
				cyan.Printf("Updating service %s\n", name)
				servicesErr := swarm.ServiceUpdate(context.Background(), currentService.ID, currentService.Version, expectedService.Spec, types.ServiceUpdateOptions{})
				if servicesErr != nil {
					return cli.NewExitError(servicesErr.Error(), 3)
				}
			} else {
				// service doesn't exist, need to create a new one
				cyan.Printf("Creating service %s\n", name)
				_, servicesErr := swarm.ServiceCreate(context.Background(), expectedService.Spec, types.ServiceCreateOptions{})
				if servicesErr != nil {
					return cli.NewExitError(servicesErr.Error(), 3)
				}
			}
		}
	}
	for name, cs := range current {
		if _, found := targetMap[cs.Spec.Name]; len(targetMap) == 0 || found {
			if _, found := expected[name]; !found {
				// service exists but it's not expected, need to delete it
				cyan.Printf("Removing service %s\n", name)
				servicesErr := swarm.ServiceRemove(context.Background(), name)
				if servicesErr != nil {
					return cli.NewExitError(servicesErr.Error(), 3)
				}
			}
		}
	}

	return nil
}

func updateNetworks(
	ctx context.Context,
	cli *client.Client,
	networks []string,
	namespace string,
) error {

	existingNetworks, err := stack.GetNetworks(ctx, cli, namespace)
	if err != nil {
		return err
	}

	existingNetworkMap := make(map[string]types.NetworkResource)
	for _, network := range existingNetworks {
		existingNetworkMap[network.Name] = network
	}

	createOpts := types.NetworkCreate{
		Labels: stack.GetStackLabels(namespace, nil),
		Driver: "overlay",
		IPAM:   network.IPAM{Driver: "default"},
	}

	for _, internalName := range networks {
		name := fmt.Sprintf("%s_%s", namespace, internalName)

		if _, exists := existingNetworkMap[name]; exists {
			continue
		}

		fmt.Printf("Creating network %s\n", name)
		if _, err := cli.NetworkCreate(ctx, name, createOpts); err != nil {
			return err
		}
	}
	return nil
}

func getUniqueNetworkNames(services map[string]bundlefile.Service) []string {
	networkSet := make(map[string]bool)
	for _, service := range services {
		for _, network := range service.Networks {
			networkSet[network] = true
		}
	}

	networks := []string{}
	for network := range networkSet {
		networks = append(networks, network)
	}
	return networks
}
