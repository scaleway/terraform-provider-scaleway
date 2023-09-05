package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultBlockTimeout = 5 * time.Minute
	blockVolumeType     = instance.VolumeServerVolumeType("sbs_volume")
)

func blockAPI(meta interface{}) *block.API {
	return block.NewAPI(meta.(*Meta).scwClient)
}

// blockAPIWithZone returns a new block API and the zone for a Create request
func blockAPIWithZone(d *schema.ResourceData, m interface{}) (*block.API, scw.Zone, error) {
	meta := m.(*Meta)
	blockAPI := block.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}

	return blockAPI, zone, nil
}

// blockAPIWithZonedAndID returns a new block API with zone and ID extracted from the state
func blockAPIWithZoneAndID(m interface{}, zonedID string) (*block.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	blockAPI := block.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(zonedID)
	if err != nil {
		return nil, "", "", err
	}

	return blockAPI, zone, ID, nil
}

func waitForBlockVolume(ctx context.Context, blockAPI *block.API, zone scw.Zone, id string, timeout time.Duration) (*block.Volume, error) {
	retryInterval := defaultFunctionRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	volume, err := blockAPI.WaitForVolumeAndReferences(&block.WaitForVolumeRequest{
		Zone:          zone,
		VolumeID:      id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return volume, err
}

// customDiffCannotShrink set key to forceNew if value shrink
func customDiffCannotShrink(key string) schema.CustomizeDiffFunc {
	return customdiff.ForceNewIf(key, func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
		oldValueI, newValueI := d.GetChange(key)
		oldValue := oldValueI.(int)
		newValue := newValueI.(int)

		return oldValue < newValue
	})
}

func waitForBlockSnapshot(ctx context.Context, blockAPI *block.API, zone scw.Zone, id string, timeout time.Duration) (*block.Snapshot, error) {
	retryInterval := defaultFunctionRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	snapshot, err := blockAPI.WaitForSnapshot(&block.WaitForSnapshotRequest{
		Zone:          zone,
		SnapshotID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return snapshot, err
}
