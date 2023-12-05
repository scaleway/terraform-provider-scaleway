package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

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
		return nil, "", err
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

func newMNQSNSAPI(d *schema.ResourceData, m any) (*mnq.SnsAPI, scw.Region, error) {
	meta := m.(*Meta)
	api := mnq.NewSnsAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func mnqSNSAPIWithRegionAndID(m interface{}, regionalID string) (*mnq.SnsAPI, scw.Region, string, error) {
	meta := m.(*Meta)
	api := mnq.NewSnsAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}
