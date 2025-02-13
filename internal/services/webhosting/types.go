package webhosting

import (
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

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
