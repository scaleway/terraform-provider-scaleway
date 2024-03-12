package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultRegistryNamespaceTimeout       = 5 * time.Minute
	defaultRegistryNamespaceRetryInterval = 5 * time.Second
)

type ErrorRegistryMessage struct {
	Error string `json:"error"`
}

// registryAPIWithRegion returns a new container registry API and the region.
func registryAPIWithRegion(d *schema.ResourceData, m interface{}) (*registry.API, scw.Region, error) {
	meta := m.(*meta.Meta)
	api := registry.NewAPI(meta.ScwClient())

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// registryAPIWithRegionAndID returns a new container registry API, region and ID.
func registryAPIWithRegionAndID(m interface{}, id string) (*registry.API, scw.Region, string, error) {
	meta := m.(*meta.Meta)
	api := registry.NewAPI(meta.ScwClient())

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func waitForRegistryNamespace(ctx context.Context, api *registry.API, region scw.Region, id string, timeout time.Duration) (*registry.Namespace, error) {
	retryInterval := defaultRegistryNamespaceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	ns, err := api.WaitForNamespace(&registry.WaitForNamespaceRequest{
		Region:        region,
		NamespaceID:   id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return ns, err
}

func waitForRegistryNamespaceDelete(ctx context.Context, api *registry.API, region scw.Region, id string, timeout time.Duration) (*registry.Namespace, error) {
	retryInterval := defaultRegistryNamespaceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	terminalStatus := map[registry.NamespaceStatus]struct{}{
		registry.NamespaceStatusReady:    {},
		registry.NamespaceStatusLocked:   {},
		registry.NamespaceStatusError:    {},
		registry.NamespaceStatusUnknown:  {},
		registry.NamespaceStatusDeleting: {},
	}

	start := time.Now()
	for {
		ns, err := api.GetNamespace(&registry.GetNamespaceRequest{
			Region:      region,
			NamespaceID: id,
		}, scw.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		if _, ok := terminalStatus[ns.Status]; ok {
			return ns, nil
		}

		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timeout while waiting for namespace %s to be deleted", id)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
		}
	}
}
