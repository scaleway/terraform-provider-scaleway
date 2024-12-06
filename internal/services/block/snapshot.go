package block

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceBlockSnapshotCreate,
		ReadContext:   ResourceBlockSnapshotRead,
		UpdateContext: ResourceBlockSnapshotUpdate,
		DeleteContext: ResourceBlockSnapshotDelete,
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
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The snapshot name",
			},
			"volume_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:      "ID of the volume from which creates a snapshot",
				DiffSuppressFunc: dsf.Locality,
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the snapshot",
			},
			"zone":       zonal.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceBlockSnapshotCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := blockAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshot, err := api.CreateSnapshot(&block.CreateSnapshotRequest{
		Zone:      zone,
		ProjectID: d.Get("project_id").(string),
		Name:      types.ExpandOrGenerateString(d.Get("name").(string), "snapshot"),
		VolumeID:  locality.ExpandID(d.Get("volume_id")),
		Tags:      types.ExpandStrings(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, snapshot.ID))

	_, err = waitForBlockSnapshot(ctx, api, zone, snapshot.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceBlockSnapshotRead(ctx, d, m)
}

func ResourceBlockSnapshotRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	snapshot, err := waitForBlockSnapshot(ctx, api, zone, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", snapshot.Name)
	_ = d.Set("zone", snapshot.Zone)
	_ = d.Set("project_id", snapshot.ProjectID)
	if snapshot.ParentVolume != nil {
		_ = d.Set("volume_id", snapshot.ParentVolume.ID)
	} else {
		_ = d.Set("volume_id", "")
	}
	_ = d.Set("tags", snapshot.Tags)

	return nil
}

func ResourceBlockSnapshotUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	snapshot, err := waitForBlockSnapshot(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	req := &block.UpdateSnapshotRequest{
		Zone:       zone,
		SnapshotID: snapshot.ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if _, err := api.UpdateSnapshot(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceBlockSnapshotRead(ctx, d, m)
}

func ResourceBlockSnapshotDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForBlockSnapshotToBeAvailable(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteSnapshot(&block.DeleteSnapshotRequest{
		Zone:       zone,
		SnapshotID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForBlockSnapshot(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
