package scaleway

import (
	"fmt"
	"strings"

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

func composeMNQID(region scw.Region, projectID string, queueName string) string {
	return fmt.Sprintf("%s/%s/%s", region, projectID, queueName)
}

func decomposeMNQID(id string) (region scw.Region, projectID string, name string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid ID format: %q", id)
	}

	region, err = scw.ParseRegion(parts[0])
	if err != nil {
		return "", "", "", err
	}

	return region, parts[1], parts[2], nil
}

func composeARN(region string, subject string, projectID string, resourceName string) string {
	return fmt.Sprintf("arn:scw:%s:%s:project-%s:%s", region, subject, projectID, resourceName)
}

func composeSNSARN(region string, projectID string, resourceName string) string {
	return composeARN(region, "sns", projectID, resourceName)
}
