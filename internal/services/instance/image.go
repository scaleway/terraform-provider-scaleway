package instance

import (
	"context"
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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceImage() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceImageCreate,
		ReadContext:   ResourceInstanceImageRead,
		UpdateContext: ResourceInstanceImageUpdate,
		DeleteContext: ResourceInstanceImageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultInstanceImageTimeout),
			Read:    schema.DefaultTimeout(defaultInstanceImageTimeout),
			Update:  schema.DefaultTimeout(defaultInstanceImageTimeout),
			Delete:  schema.DefaultTimeout(defaultInstanceImageTimeout),
			Default: schema.DefaultTimeout(defaultInstanceImageTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the image",
			},
			"root_volume_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "UUID of the snapshot from which the image is to be created",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			},
			"architecture": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          instanceSDK.ArchX86_64.String(),
				Description:      "Architecture of the image (default = x86_64)",
				ValidateDiagFunc: verify.ValidateEnum[instanceSDK.Arch](),
			},
			"additional_volume_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				},
				Description: "The IDs of the additional volumes attached to the image",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of tags [\"tag1\", \"tag2\", ...] attached to the image",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"public": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, the image will be public",
			},
			// Computed
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the image",
			},
			"modification_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last modification of the Redis cluster",
			},
			"from_server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the backed-up server from which the snapshot was taken",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The state of the image [ available | creating | error ]",
			},
			"root_volume": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Specs of the additional volumes attached to the image",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "UUID of the additional volume",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the additional volume",
							Computed:    true,
						},
						"size": {
							Type:        schema.TypeInt,
							Description: "Size of the additional volume",
							Computed:    true,
						},
						"volume_type": {
							Type:        schema.TypeString,
							Description: "Type of the additional volume",
							Computed:    true,
						},
					},
				},
			},
			"additional_volumes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Specs of the additional volumes attached to the image",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "UUID of the additional volume",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the additional volume",
							Computed:    true,
						},
						"size": {
							Type:        schema.TypeInt,
							Description: "Size of the additional volume",
							Computed:    true,
						},
						"volume_type": {
							Type:        schema.TypeString,
							Description: "Type of the additional volume",
							Computed:    true,
						},
						"tags": {
							Type:        schema.TypeList,
							Description: "List of tags attached to the additional volume",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"server": {
							Type:        schema.TypeMap,
							Description: "Server containing the volume (in case the image is a backup from a server)",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			// Common
			"zone":            zonal.Schema(),
			"project_id":      account.ProjectIDSchema(),
			"organization_id": account.OrganizationIDSchema(),
		},
		CustomizeDiff: cdf.LocalityCheck("root_volume_id", "additional_volume_ids.#"),
	}
}

func ResourceInstanceImageCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := instancehelpers.InstanceAndBlockAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instanceSDK.CreateImageRequest{
		Zone:       zone,
		Name:       types.ExpandOrGenerateString(d.Get("name"), "image"),
		RootVolume: zonal.ExpandID(d.Get("root_volume_id").(string)).ID,
		Arch:       instanceSDK.Arch(d.Get("architecture").(string)),
		Project:    types.ExpandStringPtr(d.Get("project_id")),
		Public:     types.ExpandBoolPtr(d.Get("public")),
	}

	extraVolumesIDs, volumesExist := d.GetOk("additional_volume_ids")
	if volumesExist {
		req.ExtraVolumes = expandImageExtraVolumesTemplates(locality.ExpandIDs(extraVolumesIDs))
	}

	tags, tagsExist := d.GetOk("tags")
	if tagsExist {
		req.Tags = types.ExpandStrings(tags)
	}

	if _, exist := d.GetOk("public"); exist {
		req.Public = types.ExpandBoolPtr(types.GetBool(d, "public"))
	}

	res, err := api.CreateImage(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, res.Image.ID))

	_, err = api.WaitForImage(&instanceSDK.WaitForImageRequest{
		ImageID:       res.Image.ID,
		Zone:          zone,
		RetryInterval: transport.DefaultWaitRetryInterval,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceInstanceImageRead(ctx, d, m)
}

func ResourceInstanceImageRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	image, err := instanceAPI.GetImage(&instanceSDK.GetImageRequest{
		Zone:    zone,
		ImageID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", image.Image.Name)
	_ = d.Set("root_volume_id", zonal.NewIDString(image.Image.Zone, image.Image.RootVolume.ID))
	_ = d.Set("architecture", image.Image.Arch)
	_ = d.Set("root_volume", flattenImageRootVolume(image.Image.RootVolume, zone))
	_ = d.Set("additional_volumes", flattenImageExtraVolumes(image.Image.ExtraVolumes, zone))
	_ = d.Set("tags", image.Image.Tags)
	_ = d.Set("public", image.Image.Public)
	_ = d.Set("creation_date", types.FlattenTime(image.Image.CreationDate))
	_ = d.Set("modification_date", types.FlattenTime(image.Image.ModificationDate))
	_ = d.Set("from_server_id", image.Image.FromServer)
	_ = d.Set("state", image.Image.State)
	_ = d.Set("zone", image.Image.Zone)
	_ = d.Set("project_id", image.Image.Project)
	_ = d.Set("organization_id", image.Image.Organization)

	return nil
}

func ResourceInstanceImageUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, id, err := instancehelpers.InstanceAndBlockAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instanceSDK.UpdateImageRequest{
		Zone:    zone,
		ImageID: id,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandStringPtr(d.Get("name"))
	}

	if d.HasChange("architecture") {
		req.Arch = instanceSDK.Arch(d.Get("architecture").(string))
	}

	if d.HasChange("public") {
		req.Public = types.ExpandBoolPtr(types.GetBool(d, "public"))
	}

	req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))

	image, err := api.GetImage(&instanceSDK.GetImageRequest{
		Zone:    zone,
		ImageID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("additional_volume_ids") {
		req.ExtraVolumes = expandImageExtraVolumesUpdateTemplates(locality.ExpandIDs(d.Get("additional_volume_ids")))
	} else {
		volTemplate := map[string]*instanceSDK.VolumeImageUpdateTemplate{}
		for key, vol := range image.Image.ExtraVolumes {
			volTemplate[key] = &instanceSDK.VolumeImageUpdateTemplate{
				ID: vol.ID,
			}
		}

		req.ExtraVolumes = volTemplate
	}

	// Ensure that no field is empty in request
	if req.Name == nil {
		req.Name = &image.Image.Name
	}

	if req.Arch == "" {
		req.Arch = image.Image.Arch
	}

	_, err = waitForImage(ctx, api.API, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.UpdateImage(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't update image: %w", err))
	}

	_, err = waitForImage(ctx, api.API, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceInstanceImageRead(ctx, d, m)
}

func ResourceInstanceImageDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, id, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForImage(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = instanceAPI.DeleteImage(&instanceSDK.DeleteImageRequest{
		ImageID: id,
		Zone:    zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	_, err = waitForImage(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}
