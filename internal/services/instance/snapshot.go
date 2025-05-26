package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceSnapshotCreate,
		ReadContext:   ResourceInstanceSnapshotRead,
		UpdateContext: ResourceInstanceSnapshotUpdate,
		DeleteContext: ResourceInstanceSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultInstanceSnapshotWaitTimeout),
			Delete:  schema.DefaultTimeout(defaultInstanceSnapshotWaitTimeout),
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Description:      "ID of the volume to take a snapshot from",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				ConflictsWith:    []string{"import"},
			},
			"type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				Description:      "The snapshot's volume type",
				ValidateDiagFunc: verify.ValidateEnum[instanceSDK.SnapshotVolumeType](),
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
			"import": {
				Type:     schema.TypeList,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							Description:      "Bucket containing qcow",
							DiffSuppressFunc: dsf.Locality,
							StateFunc: func(i interface{}) string {
								return regional.ExpandID(i.(string)).ID
							},
						},
						"key": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Key of the qcow file in the specified bucket",
						},
					},
				},
				Optional:      true,
				Description:   "Import snapshot from a qcow",
				ConflictsWith: []string{"volume_id"},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the snapshot",
			},
			"zone":            zonal.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
		CustomizeDiff: cdf.LocalityCheck("volume_id"),
	}
}

func ResourceInstanceSnapshotCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instanceSDK.CreateSnapshotRequest{
		Zone:    zone,
		Project: types.ExpandStringPtr(d.Get("project_id")),
		Name:    types.ExpandOrGenerateString(d.Get("name"), "snap"),
	}

	if volumeType, ok := d.GetOk("type"); ok {
		volumeType := instanceSDK.SnapshotVolumeType(volumeType.(string))
		req.VolumeType = volumeType
	}

	req.Tags = types.ExpandStringsPtr(d.Get("tags"))

	if volumeID, volumeIDExist := d.GetOk("volume_id"); volumeIDExist {
		req.VolumeID = scw.StringPtr(zonal.ExpandID(volumeID).ID)
	}

	if _, isImported := d.GetOk("import"); isImported {
		req.Bucket = types.ExpandStringPtr(regional.ExpandID(d.Get("import.0.bucket")).ID)
		req.Key = types.ExpandStringPtr(d.Get("import.0.key"))
	}

	res, err := instanceAPI.CreateSnapshot(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, res.Snapshot.ID))

	_, err = instanceAPI.WaitForSnapshot(&instanceSDK.WaitForSnapshotRequest{
		SnapshotID:    res.Snapshot.ID,
		Zone:          zone,
		RetryInterval: transport.DefaultWaitRetryInterval,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceInstanceSnapshotRead(ctx, d, m)
}

func ResourceInstanceSnapshotRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	snapshot, err := instanceAPI.GetSnapshot(&instanceSDK.GetSnapshotRequest{
		SnapshotID: id,
		Zone:       zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", snapshot.Snapshot.Name)
	_ = d.Set("created_at", snapshot.Snapshot.CreationDate.Format(time.RFC3339))
	_ = d.Set("type", snapshot.Snapshot.VolumeType.String())
	_ = d.Set("tags", snapshot.Snapshot.Tags)

	if d.Get("type").(string) == instanceSDK.VolumeVolumeTypeBSSD.String() {
		return diag.Diagnostics{
			{
				Severity:      diag.Warning,
				Summary:       "Snapshot type `b_ssd` is deprecated",
				Detail:        "If you want to migrate existing snapshots, you can visit `https://www.scaleway.com/en/docs/instances/how-to/migrate-volumes-snapshots-to-sbs/` for more information.",
				AttributePath: cty.GetAttrPath("type"),
			},
		}
	}

	return nil
}

func ResourceInstanceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instanceSDK.UpdateSnapshotRequest{
		SnapshotID: id,
		Zone:       zone,
		Name:       scw.StringPtr(d.Get("name").(string)),
		Tags:       scw.StringsPtr([]string{}),
	}

	tags := types.ExpandStrings(d.Get("tags"))
	if d.HasChange("tags") && len(tags) > 0 {
		req.Tags = scw.StringsPtr(types.ExpandStrings(d.Get("tags")))
	}

	_, err = instanceAPI.UpdateSnapshot(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't update snapshot: %w", err))
	}

	return ResourceInstanceSnapshotRead(ctx, d, m)
}

func ResourceInstanceSnapshotDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForSnapshot(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = instanceAPI.DeleteSnapshot(&instanceSDK.DeleteSnapshotRequest{
		SnapshotID: id,
		Zone:       zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	_, err = waitForSnapshot(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}
