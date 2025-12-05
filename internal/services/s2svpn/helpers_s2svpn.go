package s2svpn

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	s2svpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// NewAPIWithRegion returns a new s2svpn API and the region for a Create request
func NewAPIWithRegion(d *schema.ResourceData, m interface{}) (*s2svpn.API, scw.Region, error) {
	s2svpnAPI := s2svpn.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return s2svpnAPI, region, nil
}

// NewAPIWithRegionAndID returns a new s2svpn API with region and ID extracted from the state
func NewAPIWithRegionAndID(m interface{}, regionalID string) (*s2svpn.API, scw.Region, string, error) {
	s2svpnAPI := s2svpn.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return s2svpnAPI, region, ID, nil
}
