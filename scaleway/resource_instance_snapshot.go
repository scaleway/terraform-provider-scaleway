package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayInstanceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayInstanceSnapshotCreate,
		ReadContext:   resourceScalewayInstanceSnapshotRead,
		UpdateContext: resourceScalewayInstanceSnapshotUpdate,
		DeleteContext: resourceScalewayInstanceSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultInstanceSnapshotWaitTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the snapshot",
			},
			"volume_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "ID of the volume to take a snapshot from",
				ValidateFunc: validationUUIDorUUIDWithLocality(),
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The volume type of the snapshot",
			},
			"size_in_gb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the snapshot in gigabyte",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the snapshot",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the snapshot",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func resourceScalewayInstanceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instance.CreateSnapshotRequest{
		Zone:     zone,
		Project:  expandStringPtr(d.Get("project_id")),
		Name:     expandOrGenerateString(d.Get("name"), "snap"),
		VolumeID: expandZonedID(d.Get("volume_id").(string)).ID,
	}
	tags := expandStrings(d.Get("tags"))
	if len(tags) > 0 {
		req.Tags = tags
	}

	res, err := instanceAPI.CreateSnapshot(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.Snapshot.ID))

	_, err = instanceAPI.WaitForSnapshot(&instance.WaitForSnapshotRequest{
		SnapshotID:    res.Snapshot.ID,
		Zone:          zone,
		RetryInterval: DefaultWaitRetryInterval,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayInstanceSnapshotRead(ctx, d, meta)
}

func resourceScalewayInstanceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	snapshot, err := instanceAPI.GetSnapshot(&instance.GetSnapshotRequest{
		SnapshotID: id,
		Zone:       zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", snapshot.Snapshot.Name)
	_ = d.Set("created_at", snapshot.Snapshot.CreationDate.Format(time.RFC3339))
	_ = d.Set("type", snapshot.Snapshot.VolumeType.String())
	_ = d.Set("tags", snapshot.Snapshot.Tags)

	return nil
}

func resourceScalewayInstanceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instance.UpdateSnapshotRequest{
		SnapshotID: id,
		Zone:       zone,
		Name:       scw.StringPtr(d.Get("name").(string)),
		Tags:       scw.StringsPtr([]string{}),
	}

	tags := expandStrings(d.Get("tags"))
	if d.HasChange("tags") && len(tags) > 0 {
		req.Tags = scw.StringsPtr(expandStrings(d.Get("tags")))
	}

	_, err = instanceAPI.UpdateSnapshot(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't update snapshot: %s", err))
	}

	return resourceScalewayInstanceSnapshotRead(ctx, d, meta)
}

func resourceScalewayInstanceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForInstanceSnapshot(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = instanceAPI.DeleteSnapshot(&instance.DeleteSnapshotRequest{
		SnapshotID: id,
		Zone:       zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if !is404Error(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}
