package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// webhostingAPIWithRegion returns a new Webhosting API and the region for a Create request
func webhostingAPIWithRegion(d *schema.ResourceData, m interface{}) (*webhosting.API, scw.Region, error) {
	meta := m.(*Meta)
	api := webhosting.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
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
