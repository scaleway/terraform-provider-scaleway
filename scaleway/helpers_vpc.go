package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// vpcAPIWithZone returns a new VPC API and the zone for a Create request
func vpcAPIWithZone(d *schema.ResourceData, m interface{}) (*vpc.API, scw.Zone, error) {
	meta := m.(*Meta)
	vpcAPI := vpc.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	return vpcAPI, zone, err
}

// vpcAPIWithZoneAndID
func vpcAPIWithZoneAndID(m interface{}, id string) (*vpc.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	vpcAPI := vpc.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	return vpcAPI, zone, ID, err
}
