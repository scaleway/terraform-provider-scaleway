package marketplace

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// marketplaceAPIWithZone returns a new marketplace API and the zone for a Create request
func marketplaceAPIWithZone(d *schema.ResourceData, m interface{}) (*marketplace.API, scw.Zone, error) {
	marketplaceAPI := marketplace.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return marketplaceAPI, zone, nil
}
