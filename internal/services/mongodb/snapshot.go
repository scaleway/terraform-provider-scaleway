package mongodb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceSnapshotCreate,
		ReadContext:   ResourceSnapshotRead,
		UpdateContext: ResourceSnapshotUpdate,
		DeleteContext: ResourceSnapshotDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultMongodbSnapshotTimeout),
			Update:  schema.DefaultTimeout(defaultMongodbSnapshotTimeout),
			Delete:  schema.DefaultTimeout(defaultMongodbSnapshotTimeout),
			Default: schema.DefaultTimeout(defaultMongodbSnapshotTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Unique identifier of the snapshot",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the instance from which the snapshot was created",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the snapshot",
			},
			"instance_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the instance from which the snapshot was created",
			},

			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Size of the snapshot in bytes",
			},
			"node_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of node associated with the snapshot",
			},
			"volume_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of volume used for the snapshot (e.g., SSD, HDD)",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time when the snapshot was created",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the snapshot",
			},
			"expires_at": {
				Type:             schema.TypeString,
				Description:      "Expiration date (Format ISO 8601). Cannot be removed.",
				Required:         true,
				ValidateDiagFunc: verify.IsDate(),
			},
			"region": regional.Schema(),
		},
		CustomizeDiff: customdiff.All(),
	}
}

func ResourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	mongodbAPI, zone, region, err := newAPIWithZoneAndRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceID := locality.ExpandID(d.Get("instance_id").(string))
	createReq := &mongodb.CreateSnapshotRequest{
		InstanceID: instanceID,
		Region:     region,
		Name:       types.ExpandOrGenerateString(d.Get("name"), "snapshot"),
		ExpiresAt:  types.ExpandTimePtr(d.Get("expires_at")),
	}

	snapshot, err := mongodbAPI.CreateSnapshot(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	if snapshot != nil {
		d.SetId(zonal.NewIDString(zone, snapshot.ID))
		_, err = waitForSnapshot(ctx, mongodbAPI, region, instanceID, snapshot.ID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceSnapshotRead(ctx, d, m)
}

func ResourceSnapshotRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	mongodbAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	zone, snapshotID, err := zonal.ParseID(d.Id())

	instanceID := locality.ExpandID(d.Get("instance_id").(string))
	snapshot, err := waitForSnapshot(ctx, mongodbAPI, region, instanceID, snapshotID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("id", snapshot.ID)
	_ = d.Set("instance_id", zonal.NewIDString(zone, snapshot.InstanceID))
	_ = d.Set("name", snapshot.Name)
	_ = d.Set("instance_name", snapshot.InstanceName)
	_ = d.Set("size", snapshot.Size)
	_ = d.Set("node_type", snapshot.NodeType)
	_ = d.Set("volume_type", snapshot.VolumeType.Type)
	_ = d.Set("expires_at", types.FlattenTime(snapshot.ExpiresAt))
	_ = d.Set("created_at", types.FlattenTime(snapshot.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(snapshot.UpdatedAt))
	_ = d.Set("region", snapshot.Region.String())

	return nil
}

func ResourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	mongodbAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	_, snapshotID, err := zonal.ParseID(d.Id())

	updateReq := &mongodb.UpdateSnapshotRequest{
		SnapshotID: snapshotID,
		Region:     region,
	}

	if d.HasChange("name") {
		newName := types.ExpandOrGenerateString(d.Get("name"), "snapshot")
		updateReq.Name = &newName
	}

	if d.HasChange("expires_at") {
		updateReq.ExpiresAt = types.ExpandTimePtr(d.Get("expires_at"))
	}

	_, err = mongodbAPI.UpdateSnapshot(updateReq)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceID := locality.ExpandID(d.Get("instance_id").(string))

	_, err = waitForSnapshot(ctx, mongodbAPI, region, instanceID, snapshotID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceSnapshotRead(ctx, d, m)
}

func ResourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	mongodbAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	_, snapshotID, err := zonal.ParseID(d.Id())

	deleteReq := &mongodb.DeleteSnapshotRequest{
		SnapshotID: snapshotID,
		Region:     region,
	}

	_, err = mongodbAPI.DeleteSnapshot(deleteReq)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
