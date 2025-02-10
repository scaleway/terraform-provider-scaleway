package instance

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceVolumeCreate,
		ReadContext:   ResourceInstanceVolumeRead,
		UpdateContext: ResourceInstanceVolumeUpdate,
		DeleteContext: ResourceInstanceVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultInstanceVolumeDeleteTimeout),
			Update:  schema.DefaultTimeout(defaultInstanceVolumeDeleteTimeout),
			Delete:  schema.DefaultTimeout(defaultInstanceVolumeDeleteTimeout),
			Default: schema.DefaultTimeout(defaultInstanceVolumeDeleteTimeout),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the volume",
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "The volume type",
				ValidateDiagFunc: verify.ValidateEnum[instanceSDK.VolumeVolumeType](),
			},
			"size_in_gb": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "The size of the volume in gigabyte",
				ConflictsWith: []string{"from_snapshot_id"},
			},
			"from_snapshot_id": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Description:      "Create a volume based on a image",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				ConflictsWith:    []string{"size_in_gb"},
			},
			"server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The server associated with this volume",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the volume",
			},
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
			"zone":            zonal.Schema(),
		},
		CustomizeDiff: cdf.LocalityCheck("from_snapshot_id"),
	}
}

func ResourceInstanceVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	createVolumeRequest := &instanceSDK.CreateVolumeRequest{
		Zone:       zone,
		Name:       types.ExpandOrGenerateString(d.Get("name"), "vol"),
		VolumeType: instanceSDK.VolumeVolumeType(d.Get("type").(string)),
		Project:    types.ExpandStringPtr(d.Get("project_id")),
	}
	tags := types.ExpandStrings(d.Get("tags"))
	if len(tags) > 0 {
		createVolumeRequest.Tags = tags
	}

	if size, ok := d.GetOk("size_in_gb"); ok {
		volumeSizeInBytes := scw.Size(uint64(size.(int)) * gb)
		createVolumeRequest.Size = &volumeSizeInBytes
	}

	if snapshotID, ok := d.GetOk("from_snapshot_id"); ok {
		createVolumeRequest.BaseSnapshot = types.ExpandStringPtr(locality.ExpandID(snapshotID))
	}

	res, err := instanceAPI.CreateVolume(createVolumeRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't create volume: %s", err))
	}

	d.SetId(zonal.NewIDString(zone, res.Volume.ID))

	_, err = instanceAPI.WaitForVolume(&instanceSDK.WaitForVolumeRequest{
		VolumeID:      res.Volume.ID,
		Zone:          zone,
		RetryInterval: transport.DefaultWaitRetryInterval,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceInstanceVolumeRead(ctx, d, m)
}

func ResourceInstanceVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.GetVolume(&instanceSDK.GetVolumeRequest{
		VolumeID: id,
		Zone:     zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(fmt.Errorf("couldn't read volume: %v", err))
	}

	_ = d.Set("name", res.Volume.Name)
	_ = d.Set("organization_id", res.Volume.Organization)
	_ = d.Set("project_id", res.Volume.Project)
	_ = d.Set("zone", string(zone))
	_ = d.Set("type", res.Volume.VolumeType.String())
	_ = d.Set("tags", res.Volume.Tags)

	_, fromSnapshot := d.GetOk("from_snapshot_id")
	if !fromSnapshot {
		_ = d.Set("size_in_gb", int(res.Volume.Size/scw.GB))
	}

	if res.Volume.Server != nil {
		_ = d.Set("server_id", res.Volume.Server.ID)
	} else {
		_ = d.Set("server_id", nil)
	}

	return nil
}

func ResourceInstanceVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instanceSDK.UpdateVolumeRequest{
		VolumeID: id,
		Zone:     zone,
		Tags:     scw.StringsPtr([]string{}),
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		req.Name = &newName
	}

	tags := types.ExpandStrings(d.Get("tags"))
	if d.HasChange("tags") && len(tags) > 0 {
		req.Tags = scw.StringsPtr(types.ExpandStrings(d.Get("tags")))
	}

	if d.HasChange("size_in_gb") {
		if d.Get("type") != instanceSDK.VolumeVolumeTypeBSSD.String() {
			return diag.FromErr(errors.New("only block volume can be resized"))
		}
		if oldSize, newSize := d.GetChange("size_in_gb"); oldSize.(int) > newSize.(int) {
			return diag.FromErr(errors.New("block volumes cannot be resized down"))
		}

		_, err = waitForVolume(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		volumeSizeInBytes := scw.Size(uint64(d.Get("size_in_gb").(int)) * gb)
		_, err = instanceAPI.UpdateVolume(&instanceSDK.UpdateVolumeRequest{
			VolumeID: id,
			Zone:     zone,
			Size:     &volumeSizeInBytes,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(fmt.Errorf("couldn't resize volume: %s", err))
		}
		_, err = waitForVolume(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_, err = instanceAPI.UpdateVolume(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't update volume: %s", err))
	}

	return ResourceInstanceVolumeRead(ctx, d, m)
}

func ResourceInstanceVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	volume, err := instanceAPI.WaitForVolume(&instanceSDK.WaitForVolumeRequest{
		Zone:          zone,
		VolumeID:      id,
		RetryInterval: transport.DefaultWaitRetryInterval,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutDelete)),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	if volume.Server != nil {
		return diag.FromErr(errors.New("volume is still attached to a server"))
	}

	deleteRequest := &instanceSDK.DeleteVolumeRequest{
		Zone:     zone,
		VolumeID: id,
	}

	err = instanceAPI.DeleteVolume(deleteRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
