package tem

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	DefaultDomainTimeout           = 5 * time.Minute
	defaultDomainValidationTimeout = 60 * time.Minute
	defaultDomainRetryInterval     = 15 * time.Second
)

// temAPIWithRegion returns a new Tem API and the region for a Create request
func temAPIWithRegion(d *schema.ResourceData, m interface{}) (*tem.API, scw.Region, error) {
	api := tem.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// NewAPIWithRegionAndID returns a Tem API with zone and ID extracted from the state
func NewAPIWithRegionAndID(m interface{}, id string) (*tem.API, scw.Region, string, error) {
	api := tem.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func waitForTemDomain(ctx context.Context, api *tem.API, region scw.Region, id string, timeout time.Duration) (*tem.Domain, error) {
	retryInterval := defaultDomainRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	domain, err := api.WaitForDomain(&tem.WaitForDomainRequest{
		Region:        region,
		DomainID:      id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return domain, err
}
