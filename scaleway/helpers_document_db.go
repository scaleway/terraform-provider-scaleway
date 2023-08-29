package scaleway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	document_db "github.com/scaleway/scaleway-sdk-go/api/document_db/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultDocumentDBInstanceTimeout   = defaultRdbInstanceTimeout
	defaultWaitDocumentDBRetryInterval = defaultWaitRDBRetryInterval
)

// documentDBAPIWithRegion returns a new document_db API and the region for a Create request
func documentDBAPIWithRegion(d *schema.ResourceData, m interface{}) (*document_db.API, scw.Region, error) {
	meta := m.(*Meta)
	api := document_db.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// documentDBAPIWithRegionalAndID returns a new document_db API with region and ID extracted from the state
func documentDBAPIWithRegionAndID(m interface{}, regionalID string) (*document_db.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := document_db.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func waitForDocumentDBInstance(ctx context.Context, api *document_db.API, region scw.Region, id string, timeout time.Duration) (*document_db.Instance, error) {
	retryInterval := defaultWaitDocumentDBRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	instance, err := api.WaitForInstance(&document_db.WaitForInstanceRequest{
		Region:        region,
		InstanceID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return instance, err
}

// Build the resource identifier
// The resource identifier format is "Region/InstanceId/DatabaseName"
func resourceScalewayDocumentDBDatabaseID(region scw.Region, instanceID string, databaseName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s", region, instanceID, databaseName)
}

// resourceScalewayDocumentDBDatabaseName extract regional instanceID and databaseName from composed ID
// returned by resourceScalewayDocumentDBDatabaseID()
func resourceScalewayDocumentDBDatabaseName(id string) (string, string, error) {
	elems := strings.Split(id, "/")
	if len(elems) != 3 {
		return "", "", fmt.Errorf("cant parse terraform database id: %s", id)
	}

	return elems[0] + "/" + elems[1], elems[2], nil
}
