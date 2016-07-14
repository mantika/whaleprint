package stack

import (
	"golang.org/x/net/context"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/swarm"
)

const (
	labelNamespace = "com.docker.stack.namespace"
)

func GetStackLabels(namespace string, labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[labelNamespace] = namespace
	return labels
}

func GetStackFilter(namespace string) filters.Args {
	filter := filters.NewArgs()
	filter.Add("label", labelNamespace+"="+namespace)
	return filter
}

func GetServices(
	ctx context.Context,
	apiclient client.APIClient,
	namespace string,
) ([]swarm.Service, error) {
	return apiclient.ServiceList(
		ctx,
		types.ServiceListOptions{Filter: GetStackFilter(namespace)})
}

func GetNetworks(
	ctx context.Context,
	apiclient client.APIClient,
	namespace string,
) ([]types.NetworkResource, error) {
	return apiclient.NetworkList(
		ctx,
		types.NetworkListOptions{Filters: GetStackFilter(namespace)})
}
