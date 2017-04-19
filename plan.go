package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/client/stack"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/cli/compose/convert"
	composetypes "github.com/docker/docker/cli/compose/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

var Replica1 uint64 = 1

type Services map[string]swarm.Service

func (s Services) Keys() []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func plan(c *cli.Context) error {
	stacks, err := getStacks(c)
	if err != nil {
		return err
	}

	detail := c.Bool("detail")
	target := c.StringSlice("target")
	targetMap := map[string]bool{}

	for _, name := range target {
		targetMap[name] = true
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

		expected, err := getConfigServicesSpec(stack.Config, stack.Name, swarm)
		if err != nil {
			return cli.NewExitError(err.Error(), 3)
		}
		translateNetworkToIds(&expected, swarm, stack.Name)

		current := getSwarmServicesSpecForStack(services)

		w := bufio.NewWriter(os.Stdout)
		sp := NewServicePrinter(w, detail)

		for _, n := range expected.Keys() {
			es := expected[n]
			// Only process found target services
			if _, found := targetMap[es.Spec.Name]; len(targetMap) == 0 || found {
				if cs, found := current[n]; !found {
					// New service to create
					color.Green("+ %s", n)
					sp.PrintServiceSpec(es.Spec)
					w.Flush()
					fmt.Println()
				} else {
					different := sp.PrintServiceSpecDiff(cs.Spec, es.Spec)
					if different {
						color.Yellow("~ %s\n", es.Spec.Name)
					} else if detail {
						color.Cyan("%s\n", es.Spec.Name)
					}

					// flush if results
					if different || detail {
						w.Flush()
						fmt.Println()
					}
				}
			}
		}

		// Checks services to remove
		for _, n := range current.Keys() {
			cs := current[n]
			// Only process found target services
			if _, found := targetMap[cs.Spec.Name]; len(targetMap) == 0 || found {
				if _, found := expected[n]; !found {
					color.Red("- %s", n)
					sp.PrintServiceSpec(cs.Spec)
					w.Flush()
					fmt.Println()
				}
			}

		}
	}

	return nil
}

func translateNetworkToIds(services *Services, cli *client.Client, stackName string) {
	existingNetworks, err := stack.GetNetworks(context.Background(), cli, stackName)
	if err != nil {
		log.Fatal("Error retrieving networks")
	}

	for _, service := range *services {
		for i, network := range service.Spec.TaskTemplate.Networks {
			for _, enet := range existingNetworks {
				if enet.Name == network.Target {
					service.Spec.TaskTemplate.Networks[i].Target = enet.ID
					network.Target = enet.ID
				}
			}
		}
	}
}

func getConfigServicesSpec(config *composetypes.Config, stackName string, cli client.CommonAPIClient) (Services, error) {
	services := Services{}
	specServices, err := convert.Services(convert.NewNamespace(stackName), config, cli)
	if err != nil {
		return services, err
	}

	for n, s := range specServices {
		if s.Mode.Global == nil && s.Mode.Replicated.Replicas == nil {
			s.Mode.Replicated.Replicas = &Replica1
		}

		// hardcode VIP as it's the default service mode
		if s.EndpointSpec.Mode == "" {
			s.EndpointSpec.Mode = "vip"
		}

		if s.UpdateConfig != nil && s.UpdateConfig.FailureAction == "" {
			s.UpdateConfig.FailureAction = "pause"
		}

		fqname := fmt.Sprintf("%s_%s", stackName, n)
		s.Name = fqname
		var service = services[fqname]
		service.Spec = s
		services[fqname] = service
	}

	return services, nil

}

func getSwarmServicesSpecForStack(services []swarm.Service) Services {
	specs := map[string]swarm.Service{}

	for _, service := range services {
		specs[service.Spec.Name] = service
	}

	return specs
}
