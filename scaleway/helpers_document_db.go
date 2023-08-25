package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	document_db "github.com/scaleway/scaleway-sdk-go/api/document_db/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultDocumentDBInstanceTimeout   = defaultRdbInstanceTimeout
	defaultWaitDocumentDBRetryInterval = defaultWaitRDBRetryInterval
)

// document_dbAPIWithRegion returns a new document_db API and the region for a Create request
func document_dbAPIWithRegion(d *schema.ResourceData, m interface{}) (*document_db.API, scw.Region, error) {
	meta := m.(*Meta)
	api := document_db.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// document_dbAPIWithRegionalAndID returns a new document_db API with region and ID extracted from the state
func document_dbAPIWithRegionAndID(m interface{}, regionalID string) (*document_db.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := document_db.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func waitForDocumentDBInstance(ctx context.Context, document_dbAPI *document_db.API, region scw.Region, id string, timeout time.Duration) (*document_db.Instance, error) {
	retryInterval := defaultWaitDocumentDBRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	instance, err := document_dbAPI.WaitForInstance(&document_db.WaitForInstanceRequest{
		Region:        region,
		InstanceID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return instance, err
}
