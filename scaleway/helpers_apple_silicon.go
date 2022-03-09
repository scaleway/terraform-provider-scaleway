package scaleway

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultAppleSiliconServerTimeout = 2 * time.Minute
)

const (
	AppleSiliconM1Type = "M1-M"
)

// asAPIWithZone returns a new apple silicon API and the zone
func asAPIWithZone(d *schema.ResourceData, m interface{}) (*applesilicon.API, scw.Zone, error) {
	meta := m.(*Meta)
	asAPI := applesilicon.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return asAPI, zone, nil
}

// asAPIWithZoneAndID returns an apple silicon API with zone and ID extracted from the state
func asAPIWithZoneAndID(m interface{}, id string) (*applesilicon.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	asAPI := applesilicon.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return asAPI, zone, ID, nil
}
