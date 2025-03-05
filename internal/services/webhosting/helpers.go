package webhosting

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/webhosting/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultHostingTimeout = 5 * time.Minute
	hostingRetryInterval  = 5 * time.Second
)

func newOfferAPIWithRegion(d *schema.ResourceData, m interface{}) (*webhosting.OfferAPI, scw.Region, error) {
	api := webhosting.NewOfferAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// newHostingAPIWithRegion returns a new Hosting API and the region for a Create request.
func newHostingAPIWithRegion(d *schema.ResourceData, m interface{}) (*webhosting.HostingAPI, scw.Region, error) {
	api := webhosting.NewHostingAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func newDnsAPIWithRegion(d *schema.ResourceData, m interface{}) (*webhosting.DnsAPI, scw.Region, error) {
	api := webhosting.NewDnsAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns a Hosting API with region and ID extracted from the state.
func NewAPIWithRegionAndID(m interface{}, id string) (*webhosting.HostingAPI, scw.Region, string, error) {
	api := webhosting.NewHostingAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

func waitForHosting(ctx context.Context, api *webhosting.HostingAPI, region scw.Region, hostingID string, timeout time.Duration) (*webhosting.Hosting, error) {
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

func flattenDNSRecords(records []*webhosting.DNSRecord) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, r := range records {
		result = append(result, map[string]interface{}{
			"name":     r.Name,
			"type":     r.Type.String(),
			"ttl":      r.TTL,
			"value":    r.Value,
			"priority": r.Priority,
			"status":   r.Status.String(),
		})
	}

	return result
}

func flattenNameServers(servers []*webhosting.Nameserver) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, s := range servers {
		result = append(result, map[string]interface{}{
			"hostname":   s.Hostname,
			"status":     s.Status.String(),
			"is_default": s.IsDefault,
		})
	}

	return result
}
