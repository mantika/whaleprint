package main

import (
	"encoding/json"
	"os"

	"github.com/docker/docker/api/client/bundlefile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func export(c *cli.Context) error {
	output := c.String("output")

	swarm, swarmErr := client.NewEnvClient()
	if swarmErr != nil {
		return cli.NewExitError(swarmErr.Error(), 3)
	}

	services, servicesErr := swarm.ServiceList(context.Background(), types.ServiceListOptions{})
	if servicesErr != nil {
		return cli.NewExitError(servicesErr.Error(), 3)
	}

	dab := &bundlefile.Bundlefile{Version: "0.1", Services: map[string]bundlefile.Service{}}
	for _, service := range services {

		bundleService, err := getBundleService(service)
		if err != nil {
			return cli.NewExitError(servicesErr.Error(), 3)
		}
		dab.Services[service.Spec.Name] = *bundleService
	}

	f, err := os.Create(output)
	if err != nil {
		return cli.NewExitError(servicesErr.Error(), 3)
	}

	err = json.NewEncoder(f).Encode(dab)
	if err != nil {
		return cli.NewExitError(servicesErr.Error(), 3)
	}

	return nil
}

func getBundleService(service swarm.Service) (*bundlefile.Service, error) {
	serviceBundle := &bundlefile.Service{
		Image:      service.Spec.TaskTemplate.ContainerSpec.Image,
		Labels:     service.Spec.TaskTemplate.ContainerSpec.Labels,
		Command:    service.Spec.TaskTemplate.ContainerSpec.Command,
		Args:       service.Spec.TaskTemplate.ContainerSpec.Args,
		Env:        service.Spec.TaskTemplate.ContainerSpec.Env,
		WorkingDir: &service.Spec.TaskTemplate.ContainerSpec.Dir,
		User:       &service.Spec.TaskTemplate.ContainerSpec.User,
		Ports:      []bundlefile.Port{},
		Networks:   []string{},
	}

	for _, portcfg := range service.Endpoint.Ports {
		port := bundlefile.Port{
			Protocol:      string(portcfg.Protocol),
			Port:          portcfg.TargetPort,
			PublishedPort: portcfg.PublishedPort,
		}
		serviceBundle.Ports = append(serviceBundle.Ports, port)
	}

	for _, net := range service.Spec.Networks {
		serviceBundle.Networks = append(serviceBundle.Networks, net.Aliases...)
	}

	return serviceBundle, nil
}
