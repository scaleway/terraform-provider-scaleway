package mongodb

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultMongodbInstanceTimeout           = 30 * time.Minute
	defaultMongodbSnapshotTimeout           = 30 * time.Minute
	defaultWaitMongodbInstanceRetryInterval = 10 * time.Second
)

const (
	defaultVolumeSize = 5
)

func newAPI(m any) *mongodb.API {
	return mongodb.NewAPI(meta.ExtractScwClient(m))
}

func newAPIWithRegion(d *schema.ResourceData, m any) (*mongodb.API, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return newAPI(m), region, nil
}

// NewAPIWithRegionAndID returns a mongoDB API with region and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*mongodb.API, scw.Region, string, error) {
	region, ID, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return newAPI(m), region, ID, nil
}

func waitForInstance(ctx context.Context, api *mongodb.API, region scw.Region, id string, timeout time.Duration) (*mongodb.Instance, error) {
	retryInterval := defaultWaitMongodbInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForInstance(&mongodb.WaitForInstanceRequest{
		Timeout:       scw.TimeDurationPtr(timeout),
		InstanceID:    id,
		Region:        region,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func waitForSnapshot(ctx context.Context, api *mongodb.API, region scw.Region, instanceID string, snapshotID string, timeout time.Duration) (*mongodb.Snapshot, error) {
	retryInterval := defaultWaitMongodbInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForSnapshot(&mongodb.WaitForSnapshotRequest{
		Timeout:       scw.TimeDurationPtr(timeout),
		InstanceID:    instanceID,
		SnapshotID:    snapshotID,
		Region:        region,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}
