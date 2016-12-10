package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/client/bundlefile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func export(c *cli.Context) error {

	swarm, swarmErr := client.NewEnvClient()
	if swarmErr != nil {
		return cli.NewExitError(swarmErr.Error(), 3)
	}

	services, servicesErr := swarm.ServiceList(context.Background(), types.ServiceListOptions{})

	if len(services) == 0 {
		fmt.Println("No services found to export")
		return nil
	}

	if servicesErr != nil {
		return cli.NewExitError(servicesErr.Error(), 3)
	}

	bundles := map[string]*bundlefile.Bundlefile{}
	for _, service := range services {
		var dab *bundlefile.Bundlefile
		stackName := getStackName(service.Spec.Labels)

		if dab = bundles[stackName]; dab == nil {
			dab = &bundlefile.Bundlefile{Version: "0.1", Services: map[string]bundlefile.Service{}}
			bundles[stackName] = dab
		}

		bundleService, err := getBundleService(service)
		if err != nil {
			return cli.NewExitError(servicesErr.Error(), 3)
		}

		// Remove the stackname from the service in DAB
		service.Spec.Name = strings.TrimPrefix(service.Spec.Name, fmt.Sprintf("%s_", stackName))

		dab.Services[service.Spec.Name] = *bundleService

	}

	for output, bundle := range bundles {
		f, err := os.Create(fmt.Sprintf("%s.dab", output))
		if err != nil {
			return cli.NewExitError(servicesErr.Error(), 3)
		}

		err = json.NewEncoder(f).Encode(bundle)
		if err != nil {
			return cli.NewExitError(servicesErr.Error(), 3)
		}

		fmt.Printf("Swarm services exported successfuly for stack: %s \n", output)
		for name, _ := range bundle.Services {
			fmt.Println(name)
		}
		fmt.Println()
	}

	return nil
}

func getStackName(labels map[string]string) string {
	if stackName, ok := labels["com.docker.stack.namespace"]; ok {
		return stackName
	}
	return "services"

}

func getBundleService(service swarm.Service) (*bundlefile.Service, error) {
	serviceBundle := &bundlefile.Service{
		Image:         service.Spec.TaskTemplate.ContainerSpec.Image,
		Labels:        service.Spec.TaskTemplate.ContainerSpec.Labels,
		ServiceLabels: service.Spec.Labels,
		Command:       service.Spec.TaskTemplate.ContainerSpec.Command,
		Args:          service.Spec.TaskTemplate.ContainerSpec.Args,
		Env:           service.Spec.TaskTemplate.ContainerSpec.Env,
		WorkingDir:    &service.Spec.TaskTemplate.ContainerSpec.Dir,
		User:          &service.Spec.TaskTemplate.ContainerSpec.User,
		Ports:         []bundlefile.Port{},
		Networks:      []string{},
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
