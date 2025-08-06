package block

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
)

const (
	defaultBlockTimeout       = 5 * time.Minute
	defaultBlockRetryInterval = 5 * time.Second
	BlockVolumeType           = instance.VolumeServerVolumeType("sbs_volume")
)

// blockAPIWithZone returns a new block API and the zone for a Create request
func blockAPIWithZone(d *schema.ResourceData, m any) (*block.API, scw.Zone, error) {
	blockAPI := block.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return blockAPI, zone, nil
}

// NewAPIWithZoneAndID returns a new block API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m any, zonedID string) (*block.API, scw.Zone, string, error) {
	blockAPI := block.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(zonedID)
	if err != nil {
		return nil, "", "", err
	}

	return blockAPI, zone, ID, nil
}

// customDiffCannotShrink set key to forceNew if value shrink
func customDiffCannotShrink(key string) schema.CustomizeDiffFunc {
	return customdiff.ForceNewIf(key, func(_ context.Context, d *schema.ResourceDiff, _ any) bool {
		oldValueI, newValueI := d.GetChange(key)
		oldValue := oldValueI.(int)
		newValue := newValueI.(int)

		return oldValue > newValue
	})
}

func customDiffSnapshot(key string) schema.CustomizeDiffFunc {
	return func(ctx context.Context, diff *schema.ResourceDiff, i any) error {
		if !diff.HasChange(key) {
			return nil
		}

		oldValue, newValue := diff.GetChange(key)
		blockAPI := block.NewAPI(meta.ExtractScwClient(i))

		_, err := blockAPI.GetSnapshot(&block.GetSnapshotRequest{
			SnapshotID: oldValue.(string),
		})
		if (httperrors.Is403(err) || httperrors.Is404(err)) && newValue == nil {
			return nil
		}

		return diff.ForceNew("snapshot_id")
	}
}

func migrateInstanceToBlockVolume(ctx context.Context, api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, volumeID string, timeout time.Duration) (*block.Volume, error) {
	instanceVolumeResp, err := api.GetVolume(&instance.GetVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
	})
	if err != nil {
		return nil, err
	}

	plan, err := api.PlanBlockMigration(&instance.PlanBlockMigrationRequest{
		Zone:     instanceVolumeResp.Volume.Zone,
		VolumeID: &instanceVolumeResp.Volume.ID,
	})
	if err != nil {
		return nil, err
	}

	err = api.ApplyBlockMigration(&instance.ApplyBlockMigrationRequest{
		Zone:          instanceVolumeResp.Volume.Zone,
		VolumeID:      &instanceVolumeResp.Volume.ID,
		ValidationKey: plan.ValidationKey,
	})
	if err != nil {
		return nil, err
	}

	_, err = instancehelpers.WaitForVolume(ctx, api.API, zone, volumeID, timeout)
	if err != nil && !httperrors.Is404(err) {
		return nil, err
	}

	blockVolume, err := waitForBlockVolume(ctx, api.BlockAPI, zone, volumeID, timeout)
	if err != nil {
		return nil, err
	}

	return blockVolume, nil
}
