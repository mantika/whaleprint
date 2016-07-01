package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"sort"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/client/bundlefile"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/swarm"
	"github.com/kr/pretty"
	"github.com/urfave/cli"
)

var Replica1 uint64 = 1

type Services []swarm.ServiceSpec

func (s Services) Len() int {
	return len(s)
}
func (s Services) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
func (s Services) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func plan(c *cli.Context) error {
	stackName := c.Args().Get(0)

	if stackName == "" {
		return cli.NewExitError("Need to specify a stack name", 1)
	}
	dabLocation := c.String("dab")

	if dabLocation == "" {
		// Assume it is called as the stack name
		dabLocation = fmt.Sprintf("%s.dsb", stackName)
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

	services, servicesErr := swarm.ServiceList(context.Background(), types.ServiceListOptions{})
	if servicesErr != nil {
		return cli.NewExitError(servicesErr.Error(), 3)
	}

	desired := getBundleServicesSpec(bundle, stackName)
	current := getSwarmServicesSpecForStack(services, stackName)

	sort.Sort(desired)
	sort.Sort(current)

	log.Println(pretty.Diff(desired, current)[0])

	return nil
}

func getBundleServicesSpec(bundle *bundlefile.Bundlefile, stack string) Services {
	specs := Services{}

	for name, service := range bundle.Services {
		spec := swarm.ServiceSpec{
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: swarm.ContainerSpec{
					Image: service.Image,
				},
			},
			Networks: []swarm.NetworkAttachmentConfig{},
			Mode: swarm.ServiceMode{
				Replicated: &swarm.ReplicatedService{
					Replicas: &Replica1,
				},
			},
		}
		spec.Labels = map[string]string{"com.docker.stack.namespace": stack}
		spec.Name = fmt.Sprintf("%s_%s", stack, name)

		specs = append(specs, spec)
	}
	return specs
	/*
		Image      string
		Command    []string          `json:",omitempty"`
		Args       []string          `json:",omitempty"`
		Env        []string          `json:",omitempty"`
		Labels     map[string]string `json:",omitempty"`
		Ports      []Port            `json:",omitempty"`
		WorkingDir *string           `json:",omitempty"`
		User       *string           `json:",omitempty"`
		Networks   []string          `json:",omitempty"`
	*/
}

func getSwarmServicesSpecForStack(services []swarm.Service, stack string) Services {
	specs := Services{}

	for _, service := range services {
		log.Println(service.Spec.Labels["com.docker.stack.namespace"])
		if service.Spec.Labels["com.docker.stack.namespace"] == stack {
			specs = append(specs, service.Spec)
		}
	}

	return specs
}
