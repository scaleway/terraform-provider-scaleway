package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// marketplaceAPIWithZone returns a new marketplace API and the zone for a Create request
func marketplaceAPIWithZone(d *schema.ResourceData, m interface{}) (*marketplace.API, scw.Zone, error) {
	meta := m.(*Meta)
	marketplaceAPI := marketplace.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	return marketplaceAPI, zone, err
}
