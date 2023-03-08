package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func newMNQAPI(d *schema.ResourceData, m interface{}) (*mnq.API, scw.Region, error) {
	meta := m.(*Meta)
	api := mnq.NewAPI(meta.scwClient)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func mnqAPIWithRegionAndID(m interface{}, regionalID string) (*mnq.API, scw.Region, string, error) {
	meta := m.(*Meta)
	mnqAPI := mnq.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}
	return mnqAPI, region, ID, nil
}
