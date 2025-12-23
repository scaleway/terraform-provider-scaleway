package block

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceBlockVolumeCreate,
		ReadContext:   ResourceBlockVolumeRead,
		UpdateContext: ResourceBlockVolumeUpdate,
		DeleteContext: ResourceBlockVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultBlockTimeout),
			Read:    schema.DefaultTimeout(defaultBlockTimeout),
			Delete:  schema.DefaultTimeout(defaultBlockTimeout),
			Default: schema.DefaultTimeout(defaultBlockTimeout),
		},
		SchemaVersion: 0,
		Identity:      identity.DefaultZonal(),
		SchemaFunc:    volumeSchema,
		CustomizeDiff: customdiff.All(
			customDiffSnapshot("snapshot_id"),
			customDiffCannotShrink("size_in_gb"),
		),
	}
}

func volumeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The volume name",
		},
		"iops": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "The maximum IO/s expected, must match available options",
		},
		"size_in_gb": {
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			Description: "The volume size in GB",
		},
		"snapshot_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The snapshot to create the volume from",
			DiffSuppressFunc: dsf.Locality,
		},
		"instance_volume_id": {
			Type:          schema.TypeString,
			Computed:      true,
			Optional:      true,
			Description:   "The instance volume to create the block volume from",
			ForceNew:      true,
			ConflictsWith: []string{"snapshot_id"},
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The tags associated with the volume",
		},
		"zone":       zonal.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func ResourceBlockVolumeCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := instancehelpers.InstanceAndBlockAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var volume *block.Volume

	instanceVolumeID, migrateVolume := d.GetOk("instance_volume_id")
	if migrateVolume {
		volume, err = migrateInstanceToBlockVolume(ctx, api, zone, locality.ExpandID(instanceVolumeID.(string)), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		req := &block.CreateVolumeRequest{
			Zone:      zone,
			Name:      types.ExpandOrGenerateString(d.Get("name").(string), "volume"),
			ProjectID: d.Get("project_id").(string),
			Tags:      types.ExpandStrings(d.Get("tags")),
			PerfIops:  types.ExpandUint32Ptr(d.Get("iops")),
		}

		if iops, ok := d.GetOk("iops"); ok {
			req.PerfIops = types.ExpandUint32Ptr(iops)
		}

		snapshotID, hasSnapshot := d.GetOk("snapshot_id")
		if hasSnapshot {
			req.FromSnapshot = &block.CreateVolumeRequestFromSnapshot{
				SnapshotID: locality.ExpandID(snapshotID.(string)),
			}
		}

		if size, ok := d.GetOk("size_in_gb"); ok {
			volumeSizeInBytes := scw.Size(size.(int)) * scw.GB
			if hasSnapshot {
				req.FromSnapshot.Size = &volumeSizeInBytes
			} else {
				req.FromEmpty = &block.CreateVolumeRequestFromEmpty{
					Size: volumeSizeInBytes,
				}
			}
		}

		volume, err = api.BlockAPI.CreateVolume(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(zonal.NewIDString(zone, volume.ID))

	_, err = waitForBlockVolume(ctx, api.BlockAPI, zone, volume.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceBlockVolumeRead(ctx, d, m)
}

func ResourceBlockVolumeRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	volume, err := waitForBlockVolume(ctx, api, zone, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", volume.Name)

	if volume.Specs != nil {
		_ = d.Set("iops", types.FlattenUint32Ptr(volume.Specs.PerfIops))
	}

	_ = d.Set("size_in_gb", int(volume.Size/scw.GB))
	_ = d.Set("zone", volume.Zone)
	_ = d.Set("project_id", volume.ProjectID)
	_ = d.Set("tags", volume.Tags)
	snapshotID := ""

	if volume.ParentSnapshotID != nil {
		_, err := api.GetSnapshot(&block.GetSnapshotRequest{
			SnapshotID: *volume.ParentSnapshotID,
			Zone:       zone,
		})

		if err == nil || (!httperrors.Is403(err) && !httperrors.Is404(err)) {
			snapshotID = zonal.NewIDString(zone, *volume.ParentSnapshotID)
		}
	}

	_ = d.Set("snapshot_id", snapshotID)

	return nil
}

func ResourceBlockVolumeUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	volume, err := waitForBlockVolume(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	req := &block.UpdateVolumeRequest{
		Zone:     zone,
		VolumeID: volume.ID,
		Name:     types.ExpandUpdatedStringPtr(volume.Name),
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("size_in_gb") {
		volumeSizeInBytes := scw.Size(uint64(d.Get("size_in_gb").(int)) * gb)
		req.Size = &volumeSizeInBytes
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("iops") {
		req.PerfIops = types.ExpandUint32Ptr(d.Get("iops"))
	}

	if _, err := api.UpdateVolume(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceBlockVolumeRead(ctx, d, m)
}

func ResourceBlockVolumeDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForBlockVolumeToBeAvailable(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	err = api.DeleteVolume(&block.DeleteVolumeRequest{
		Zone:     zone,
		VolumeID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForBlockVolume(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
