package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func resourceScalewayInstanceImage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayInstanceImageCreate,
		ReadContext:   resourceScalewayInstanceImageRead,
		UpdateContext: resourceScalewayInstanceImageUpdate,
		DeleteContext: resourceScalewayInstanceImageDelete,
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
				Type:         schema.TypeString,
				Required:     true,
				Description:  "UUID of the snapshot from which the image is to be created",
				ValidateFunc: validationUUIDorUUIDWithLocality(),
			},
			"architecture": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     instance.ArchX86_64.String(),
				Description: "Architecture of the image (default = x86_64)",
				ValidateFunc: validation.StringInSlice([]string{
					instance.ArchArm.String(),
					instance.ArchX86_64.String(),
				}, false),
			},
			"additional_volume_ids": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validationUUIDorUUIDWithLocality(),
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
			"additional_volumes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Specs of the additional volumes attached to the image",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"export_uri": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creation_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"modification_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"organization": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"project": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"server": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			// Common
			"zone":            zonal.Schema(),
			"project_id":      projectIDSchema(),
			"organization_id": organizationIDSchema(),
		},
		CustomizeDiff: CustomizeDiffLocalityCheck("root_volume_id", "additional_volume_ids.#"),
	}
}

func resourceScalewayInstanceImageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instance.CreateImageRequest{
		Zone:       zone,
		Name:       expandOrGenerateString(d.Get("name"), "image"),
		RootVolume: zonal.ExpandID(d.Get("root_volume_id").(string)).ID,
		Arch:       instance.Arch(d.Get("architecture").(string)),
		Project:    expandStringPtr(d.Get("project_id")),
		Public:     expandBoolPtr(d.Get("public")),
	}

	extraVolumesIDs, volumesExist := d.GetOk("additional_volume_ids")
	if volumesExist {
		snapResponses, err := getSnapshotsFromIDs(ctx, extraVolumesIDs.([]interface{}), instanceAPI)
		if err != nil {
			return diag.FromErr(err)
		}
		req.ExtraVolumes = expandInstanceImageExtraVolumesTemplates(snapResponses)
	}
	tags, tagsExist := d.GetOk("tags")
	if tagsExist {
		req.Tags = expandStrings(tags)
	}
	if _, exist := d.GetOk("public"); exist {
		req.Public = expandBoolPtr(getBool(d, "public"))
	}

	res, err := instanceAPI.CreateImage(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, res.Image.ID))

	_, err = instanceAPI.WaitForImage(&instance.WaitForImageRequest{
		ImageID:       res.Image.ID,
		Zone:          zone,
		RetryInterval: transport.DefaultWaitRetryInterval,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayInstanceImageRead(ctx, d, m)
}

func resourceScalewayInstanceImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	image, err := instanceAPI.GetImage(&instance.GetImageRequest{
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
	_ = d.Set("additional_volumes", flattenInstanceImageExtraVolumes(image.Image.ExtraVolumes, zone))
	_ = d.Set("tags", image.Image.Tags)
	_ = d.Set("public", image.Image.Public)
	_ = d.Set("creation_date", flattenTime(image.Image.CreationDate))
	_ = d.Set("modification_date", flattenTime(image.Image.ModificationDate))
	_ = d.Set("from_server_id", image.Image.FromServer)
	_ = d.Set("state", image.Image.State)
	_ = d.Set("zone", image.Image.Zone)
	_ = d.Set("project_id", image.Image.Project)
	_ = d.Set("organization_id", image.Image.Organization)

	return nil
}

func resourceScalewayInstanceImageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instance.UpdateImageRequest{
		Zone:    zone,
		ImageID: id,
	}

	if d.HasChange("name") {
		req.Name = expandStringPtr(d.Get("name"))
	}
	if d.HasChange("architecture") {
		req.Arch = instance.Arch(d.Get("architecture").(string))
	}
	if d.HasChange("public") {
		req.Public = expandBoolPtr(getBool(d, "public"))
	}
	req.Tags = expandUpdatedStringsPtr(d.Get("tags"))

	image, err := instanceAPI.GetImage(&instance.GetImageRequest{
		Zone:    zone,
		ImageID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("additional_volume_ids") {
		snapResponses, err := getSnapshotsFromIDs(ctx, d.Get("additional_volume_ids").([]interface{}), instanceAPI)
		if err != nil {
			return diag.FromErr(err)
		}
		req.ExtraVolumes = expandInstanceImageExtraVolumesUpdateTemplates(snapResponses)
	} else {
		volTemplate := map[string]*instance.VolumeImageUpdateTemplate{}
		for key, vol := range image.Image.ExtraVolumes {
			volTemplate[key] = &instance.VolumeImageUpdateTemplate{
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

	_, err = waitForInstanceImage(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = instanceAPI.UpdateImage(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("couldn't update image: %s", err))
	}

	_, err = waitForInstanceImage(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayInstanceImageRead(ctx, d, m)
}

func resourceScalewayInstanceImageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForInstanceImage(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = instanceAPI.DeleteImage(&instance.DeleteImageRequest{
		ImageID: id,
		Zone:    zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	_, err = waitForInstanceImage(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}
