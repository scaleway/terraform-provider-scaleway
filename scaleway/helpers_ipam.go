package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ipam "github.com/scaleway/scaleway-sdk-go/api/ipam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// ipamAPIWithRegion returns a new ipam API and the region
func ipamAPIWithRegion(d *schema.ResourceData, m interface{}) (*ipam.API, scw.Region, error) {
	meta := m.(*Meta)
	ipamAPI := ipam.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return ipamAPI, region, nil
}
