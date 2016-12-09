package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/client/bundlefile"
	"github.com/docker/docker/api/client/stack"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
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

		expected := getBundleServicesSpec(stack.Bundle, stack.Name)
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

func safeDereference(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func translateNetworkToIds(services *Services, cli *client.Client, stackName string) {
	existingNetworks, err := stack.GetNetworks(context.Background(), cli, stackName)
	if err != nil {
		log.Fatal("Error retrieving networks")
	}

	for _, service := range *services {
		for i, network := range service.Spec.Networks {
			for _, enet := range existingNetworks {
				if enet.Name == network.Target {
					service.Spec.Networks[i].Target = enet.ID
					network.Target = enet.ID
				}
			}
		}
	}
}

func getBundleServicesSpec(bundle *bundlefile.Bundlefile, stackName string) Services {
	specs := Services{}

	for name, service := range bundle.Services {
		spec := swarm.ServiceSpec{
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: swarm.ContainerSpec{
					Image:   service.Image,
					Labels:  service.Labels,
					Command: service.Command,
					Args:    service.Args,
					Env:     service.Env,
					Dir:     safeDereference(service.WorkingDir),
					User:    safeDereference(service.User),
				},
				Placement: &swarm.Placement{Constraints: service.Constraints},
			},
			Networks: convertNetworks(service.Networks, stackName, name),
		}

		spec.Mode = getServiceMode(service.Mode)

		if service.Replicas != nil {
			spec.Mode.Replicated.Replicas = service.Replicas
		}

		spec.Labels = map[string]string{"com.docker.stack.namespace": stackName}
		spec.Name = fmt.Sprintf("%s_%s", stackName, name)

		// Populate ports
		ports := []swarm.PortConfig{}
		for _, port := range service.Ports {
			p := swarm.PortConfig{
				TargetPort:    port.Port,
				Protocol:      swarm.PortConfigProtocol(port.Protocol),
				PublishedPort: port.PublishedPort,
			}

			ports = append(ports, p)
		}
		// Hardcode resolution mode to VIP as it's the default with dab
		mode := "vip"
		if service.EndpointMode != nil {
			if *service.EndpointMode != "dnsrr" && *service.EndpointMode != "vip" {
				log.Fatalf("Invalid mode \"%s\" for service %s, only \"dnsrr\" or \"vip\" is allowed", *service.EndpointMode, spec.Name)
			}
			mode = *service.EndpointMode
		}
		spec.EndpointSpec = &swarm.EndpointSpec{Ports: ports, Mode: swarm.ResolutionMode(mode)}

		service := swarm.Service{}
		service.ID = spec.Name
		service.Spec = spec

		specs[spec.Name] = service
	}
	return specs
}

func getServiceMode(mode *string) swarm.ServiceMode {
	if mode != nil && *mode == "global" {
		return swarm.ServiceMode{
			Global: &swarm.GlobalService{},
		}
	} else {
		return swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &Replica1,
			},
		}
	}
}

func convertNetworks(networks []string, namespace string, name string) []swarm.NetworkAttachmentConfig {
	nets := []swarm.NetworkAttachmentConfig{}
	for _, network := range networks {
		nets = append(nets, swarm.NetworkAttachmentConfig{
			Target:  namespace + "_" + network,
			Aliases: []string{name},
		})
	}
	return nets
}

func getSwarmServicesSpecForStack(services []swarm.Service) Services {
	specs := Services{}

	for _, service := range services {
		specs[service.Spec.Name] = service
	}

	return specs
}
