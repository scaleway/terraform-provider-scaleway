package mongodb

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultMongodbInstanceTimeout           = 15 * time.Minute
	defaultMongodbSnapshotTimeout           = 15 * time.Minute
	defaultWaitMongodbInstanceRetryInterval = 5 * time.Second
)

const (
	defaultVolumeSize = 5
)

func newAPI(m interface{}) *mongodb.API {
	return mongodb.NewAPI(meta.ExtractScwClient(m))
}

// newAPIWithZone returns a new mongoDB API and the zone for a Create request
func newAPIWithZone(d *schema.ResourceData, m interface{}) (*mongodb.API, scw.Zone, error) {
	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}
	return newAPI(m), zone, nil
}

func newAPIWithZoneAndRegion(d *schema.ResourceData, m interface{}) (*mongodb.API, scw.Zone, scw.Region, error) {
	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", "", err
	}
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", "", err
	}
	return newAPI(m), zone, region, nil
}

func newAPIWithRegion(d *schema.ResourceData, m interface{}) (*mongodb.API, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}
	return newAPI(m), region, nil
}

// NewAPIWithZoneAndID returns a mongoDB API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m interface{}, id string) (*mongodb.API, scw.Zone, string, error) {
	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	return newAPI(m), zone, ID, nil
}

func NewAPIWithRegionAndID(m interface{}, id string) (*mongodb.API, scw.Region, string, error) {
	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}
	region, err := zone.Region()
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
