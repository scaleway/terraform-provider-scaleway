package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayBlockVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayBlockVolumeCreate,
		ReadContext:   resourceScalewayBlockVolumeRead,
		UpdateContext: resourceScalewayBlockVolumeUpdate,
		DeleteContext: resourceScalewayBlockVolumeDelete,
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
				Description: "The volume name",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The volume type",
			},
			"size_in_gb": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "The volume size in GB",
				ExactlyOneOf: []string{"snapshot_id"},
			},
			"snapshot_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The snapshot to create the volume from",
				ExactlyOneOf: []string{"size_in_gb"},
			},
			"snapshot_project_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "ID of the project where the snapshot is",
				RequiredWith: []string{"snapshot_id"},
			},
			"zone":       zoneSchema(),
			"project_id": projectIDSchema(),
		},
		CustomizeDiff: customdiff.All(
			customDiffCannotShrink("size_in_gb"),
		),
	}
}

func resourceScalewayBlockVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, zone, err := blockAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &block.CreateVolumeRequest{
		Zone:      zone,
		Name:      expandOrGenerateString(d.Get("name").(string), "volume"),
		ProjectID: d.Get("project_id").(string),
		Type:      d.Get("type").(string),
	}

	if size, ok := d.GetOk("size_in_gb"); ok {
		volumeSizeInBytes := scw.Size(uint64(size.(int)) * gb)
		req.Size = &volumeSizeInBytes
	}

	if snapshotID, ok := d.GetOk("snapshot_id"); ok {
		req.Snapshot = &block.SnapshotSpec{
			SnapshotID: snapshotID.(string),
			ProjectID:  expandStringPtr(d.Get("snapshot_project_id")),
		}
	}

	volume, err := api.CreateVolume(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, volume.ID))

	_, err = waitForBlockVolume(ctx, api, zone, volume.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayBlockVolumeRead(ctx, d, meta)
}

func resourceScalewayBlockVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, zone, id, err := blockAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	volume, err := waitForBlockVolume(ctx, api, zone, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", volume.Name)
	_ = d.Set("type", volume.Type)
	_ = d.Set("size_in_gb", volume.Size/scw.GB)
	_ = d.Set("snapshot_id", flattenStringPtr(volume.ParentSnapshotID))
	_ = d.Set("zone", volume.Zone)
	_ = d.Set("project_id", volume.ProjectID)

	return nil
}

func resourceScalewayBlockVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, zone, id, err := blockAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	volume, err := waitForBlockVolume(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	req := &block.UpdateVolumeRequest{
		Zone:     zone,
		VolumeID: volume.ID,
	}

	if d.HasChange("name") {
		req.Name = expandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("size") {
		volumeSizeInBytes := scw.Size(uint64(d.Get("size").(int)) * gb)
		req.Size = &volumeSizeInBytes
	}

	if _, err := api.UpdateVolume(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayBlockVolumeRead(ctx, d, meta)
}

func resourceScalewayBlockVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, zone, id, err := blockAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForBlockVolume(ctx, api, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
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
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
