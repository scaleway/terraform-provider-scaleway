package scaleway

import (
	"github.com/hashicorp/terraform/helper/schema"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// getLbAPI returns a new lb API
func getLbAPI(m interface{}) *lb.API {
	meta := m.(*Meta)
	return lb.NewAPI(meta.scwClient)
}

// getLbAPIWithRegion returns a new lb API and the region for a Create request
func getLbAPIWithRegion(d *schema.ResourceData, m interface{}) (*lb.API, scw.Region, error) {
	meta := m.(*Meta)
	lbApi := lb.NewAPI(meta.scwClient)

	region, err := getRegion(d, meta)
	return lbApi, region, err
}

// getLbAPIWithRegionAndID returns an lb API with region and ID extracted from the state
func getLbAPIWithRegionAndID(m interface{}, id string) (*lb.API, scw.Region, string, error) {
	meta := m.(*Meta)
	lbApi := lb.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(id)
	return lbApi, region, ID, err
}
