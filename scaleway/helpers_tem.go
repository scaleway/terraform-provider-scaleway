package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultTemDomainTimeout       = 5 * time.Minute
	defaultTemDomainRetryInterval = 15 * time.Second
)

type ErrorTemMessage struct {
	Error string
}

// temAPIWithRegion returns a new Tem API and the region for a Create request
func temAPIWithRegion(d *schema.ResourceData, m interface{}) (*tem.API, scw.Region, error) {
	meta := m.(*Meta)
	api := tem.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// temAPIWithRegionAndID returns a Tem API with zone and ID extracted from the state
func temAPIWithRegionAndID(m interface{}, id string) (*tem.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := tem.NewAPI(meta.scwClient)

	region, id, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func waitForTemDomain(ctx context.Context, api *tem.API, region scw.Region, id string, timeout time.Duration) (*tem.Domain, error) {
	retryInterval := defaultRegistryNamespaceRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	domain, err := api.WaitForDomain(&tem.WaitForDomainRequest{
		Region:        region,
		DomainID:      id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return domain, err
}
