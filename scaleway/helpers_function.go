package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultFunctionNamespaceTimeout = 5 * time.Minute
	defaultFunctionRetryInterval    = 5 * time.Second
)

// functionAPIWithRegion returns a new container registry API and the region.
func functionAPIWithRegion(d *schema.ResourceData, m interface{}) (*function.API, scw.Region, error) {
	meta := m.(*Meta)
	api := function.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// functionAPIWithRegionAndID returns a new container registry API, region and ID.
func functionAPIWithRegionAndID(m interface{}, id string) (*function.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := function.NewAPI(meta.scwClient)

	region, id, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func waitForFunctionNamespace(ctx context.Context, d *schema.ResourceData, meta interface{}) (*function.Namespace, error) {
	api, region, id, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return nil, err
	}

	retryInterval := defaultFunctionRetryInterval

	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	ns, err := api.WaitForNamespace(&function.WaitForNamespaceRequest{
		Region:        region,
		NamespaceID:   id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(defaultFunctionNamespaceTimeout),
	}, scw.WithContext(ctx))

	return ns, err
}
