package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnqAlpha "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func newMNQAPIalpha(d *schema.ResourceData, m interface{}) (*mnqAlpha.API, scw.Region, error) {
	meta := m.(*Meta)
	api := mnqAlpha.NewAPI(meta.scwClient)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func mnqAPIWithRegionAndIDalpha(m interface{}, regionalID string) (*mnqAlpha.API, scw.Region, string, error) {
	meta := m.(*Meta)
	mnqAPI := mnqAlpha.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}
	return mnqAPI, region, ID, nil
}

func newMNQNatsAPI(d *schema.ResourceData, m interface{}) (*mnq.NatsAPI, scw.Region, error) {
	meta := m.(*Meta)
	api := mnq.NewNatsAPI(meta.scwClient)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func mnqNatsAPIWithRegionAndID(m interface{}, regionalID string) (*mnq.NatsAPI, scw.Region, string, error) {
	meta := m.(*Meta)
	api := mnq.NewNatsAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, ID, nil
}

func newMNQSQSAPI(d *schema.ResourceData, m any) (*mnq.SqsAPI, scw.Region, error) {
	meta := m.(*Meta)
	api := mnq.NewSqsAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", nil
	}

	return api, region, nil
}
func mnqSQSAPIWithRegionAndID(m interface{}, regionalID string) (*mnq.SqsAPI, scw.Region, string, error) {
	meta := m.(*Meta)
	api := mnq.NewSqsAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, ID, nil
}
