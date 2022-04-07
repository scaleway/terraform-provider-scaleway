package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultRegistryNamespaceTimeout       = 5 * time.Minute
	defaultRegistryNamespaceRetryInterval = 5 * time.Second
)

type ErrorRegistryMessage struct {
	Error string
}

// registryAPIWithRegion returns a new container registry API and the region.
func registryAPIWithRegion(d *schema.ResourceData, m interface{}) (*registry.API, scw.Region, error) {
	meta := m.(*Meta)
	api := registry.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// registryAPIWithRegionAndID returns a new container registry API, region and ID.
func registryAPIWithRegionAndID(m interface{}, id string) (*registry.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := registry.NewAPI(meta.scwClient)

	region, id, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func waitForRegistryNamespace(ctx context.Context, d *schema.ResourceData, meta interface{}, timeout time.Duration) (*registry.Namespace, error) {
	api, region, id, err := registryAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return nil, err
	}

	retryInterval := defaultRegistryNamespaceRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	ns, err := api.WaitForNamespace(&registry.WaitForNamespaceRequest{
		Region:        region,
		NamespaceID:   id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return ns, err
}
