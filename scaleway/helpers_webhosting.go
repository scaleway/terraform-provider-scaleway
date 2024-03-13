package scaleway

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
	defaultWebhostingTimeout = 5 * time.Minute
	hostingRetryInterval     = 5 * time.Second
)

// webhostingAPIWithRegion returns a new Webhosting API and the region for a Create request
func webhostingAPIWithRegion(d *schema.ResourceData, m interface{}) (*webhosting.API, scw.Region, error) {
	api := webhosting.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// webhostingAPIWithRegionAndID returns a Webhosting API with region and ID extracted from the state
func webhostingAPIWithRegionAndID(m interface{}, id string) (*webhosting.API, scw.Region, string, error) {
	api := webhosting.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func flattenOfferProduct(product *webhosting.OfferProduct) interface{} {
	return []map[string]interface{}{
		{
			"name":                  product.Name,
			"option":                product.Option,
			"email_accounts_quota":  product.EmailAccountsQuota,
			"email_storage_quota":   product.EmailStorageQuota,
			"databases_quota":       product.DatabasesQuota,
			"hosting_storage_quota": product.HostingStorageQuota,
			"support_included":      product.SupportIncluded,
			"v_cpu":                 product.VCPU,
			"ram":                   product.RAM,
		},
	}
}

func flattenOfferPrice(price *scw.Money) interface{} {
	return price.String()
}

func flattenHostingCpanelUrls(cpanelURL *webhosting.HostingCpanelURLs) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"dashboard": cpanelURL.Dashboard,
			"webmail":   cpanelURL.Webmail,
		},
	}
}

func flattenHostingOptions(options []*webhosting.HostingOption) []map[string]interface{} {
	if options == nil {
		return nil
	}
	flattenedOptions := []map[string]interface{}(nil)
	for _, option := range options {
		flattenedOptions = append(flattenedOptions, map[string]interface{}{
			"id":   option.ID,
			"name": option.Name,
		})
	}
	return flattenedOptions
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
