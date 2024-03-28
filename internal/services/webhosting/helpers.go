package webhosting

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultHostingTimeout = 5 * time.Minute
	hostingRetryInterval  = 5 * time.Second
)

// newAPIWithRegion returns a new Webhosting API and the region for a Create request
func newAPIWithRegion(d *schema.ResourceData, m interface{}) (*webhosting.API, scw.Region, error) {
	api := webhosting.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// NewAPIWithRegionAndID returns a Webhosting API with region and ID extracted from the state
func NewAPIWithRegionAndID(m interface{}, id string) (*webhosting.API, scw.Region, string, error) {
	api := webhosting.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func waitForHosting(ctx context.Context, api *webhosting.API, region scw.Region, hostingID string, timeout time.Duration) (*webhosting.Hosting, error) {
	retryInterval := hostingRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForHosting(&webhosting.WaitForHostingRequest{
		HostingID:     hostingID,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}
