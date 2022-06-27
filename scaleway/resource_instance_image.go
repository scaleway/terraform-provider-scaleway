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
				Required:    true,
				Description: "Architecture of the image",
			},
			"default_bootscript_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "ID of the default bootscript of the image",
				ValidateFunc: validationUUIDorUUIDWithLocality(),
			},
			"additional_volume_ids": {
				Type:     schema.TypeList,
				Optional: true,
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
			"default_bootscript": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Specs of the default bootscript of the image",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				//Elem: &schema.Resource{
				//	Schema: map[string]*schema.Schema{
				//		"bootcmdargs": {
				//			Type: schema.TypeString,
				//		},
				//		"default": {
				//			Type: schema.TypeString,
				//		},
				//		"dtb": {
				//			Type: schema.TypeString,
				//		},
				//		"id": {
				//			Type: schema.TypeString,
				//		},
				//		"initrd": {
				//			Type: schema.TypeString,
				//		},
				//		"kernel": {
				//			Type: schema.TypeString,
				//		},
				//		"organization": {
				//			Type: schema.TypeString,
				//		},
				//		"project": {
				//			Type: schema.TypeString,
				//		},
				//		"public": {
				//			Type: schema.TypeString,
				//		},
				//		"title": {
				//			Type: schema.TypeString,
				//		},
				//		"arch": {
				//			Type: schema.TypeString,
				//		},
				//		"zone": {
				//			Type: schema.TypeString,
				//		},
				//	},
				//},
			},
			"additional_volumes": {
				Type:        schema.TypeSet,
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
							//Elem: &schema.Resource{
							//	Schema: map[string]*schema.Schema{
							//		"id": {
							//			Type: schema.TypeString,
							//		},
							//		"name": {
							//			Type: schema.TypeString,
							//		},
							//	},
							//},
						},
					},
				},
			},
			// Common
			"zone":            zoneSchema(),
			"project_id":      projectIDSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayInstanceImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instance.CreateImageRequest{
		Zone:       zone,
		Name:       expandOrGenerateString(d.Get("name"), "image"),
		RootVolume: expandZonedID(d.Get("root_volume_id").(string)).ID,
		Arch:       instance.Arch(d.Get("architecture").(string)),
		Project:    expandStringPtr(d.Get("project_id")),
		Public:     false,
	}

	defaultBootscript, bootscriptExists := d.GetOk("default_bootscript_id")
	if bootscriptExists {
		req.DefaultBootscript = defaultBootscript.(string)
	}
	extraVolumesIds, volumesExist := d.GetOk("additional_volume_ids")
	if volumesExist {
		snapResponses, err := getExtraVolumesSpecsFromSnapshots(extraVolumesIds.([]interface{}), instanceAPI, ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		req.ExtraVolumes = expandInstanceImageExtraVolumesTemplates(snapResponses)
	}
	tags, tagsExist := d.GetOk("tags")
	if tagsExist {
		req.Tags = expandStrings(tags)
	}
	if isPublic := d.Get("public"); isPublic == true {
		req.Public = true
	}

	res, err := instanceAPI.CreateImage(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.Image.ID))

	_, err = instanceAPI.WaitForImage(&instance.WaitForImageRequest{
		ImageID:       res.Image.ID,
		Zone:          zone,
		RetryInterval: DefaultWaitRetryInterval,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayInstanceImageRead(ctx, d, meta)
}

func resourceScalewayInstanceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	image, err := instanceAPI.GetImage(&instance.GetImageRequest{
		Zone:    zone,
		ImageID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", image.Image.Name)
	_ = d.Set("root_volume_id", newZonedIDString(image.Image.Zone, image.Image.RootVolume.ID))
	_ = d.Set("architecture", image.Image.Arch)
	if _, defaultBootscriptExists := d.GetOk("default_bootscript_id"); defaultBootscriptExists == true {
		_ = d.Set("default_bootscript", flattenInstanceImageBootscript(image.Image.DefaultBootscript))
	}
	if _, extraVolumesExist := d.GetOk("additional_volume_ids"); extraVolumesExist == true {
		//additionalVolumeIDs := []string(nil)
		//for _, volume := range orderVolumes(image.Image.ExtraVolumes) {
		//	additionalVolumeIDs = append(additionalVolumeIDs, volume.ID)
		//}
		//_ = d.Set("additional_volume_ids", additionalVolumeIDs)
		//TODO: do i need to list all volume specs ? if not then the commented code above will suffice
		_ = d.Set("additional_volumes", flattenInstanceImageExtraVolumes(image.Image.ExtraVolumes))
	}
	_ = d.Set("tags", image.Image.Tags)
	_ = d.Set("public", image.Image.Public)
	_ = d.Set("creation_date", image.Image.CreationDate.Format(time.RFC3339))
	_ = d.Set("modification_date", image.Image.ModificationDate.Format(time.RFC3339))
	_ = d.Set("from_server_id", image.Image.FromServer)
	_ = d.Set("state", image.Image.State)
	_ = d.Set("zone", image.Image.Zone)
	_ = d.Set("project_id", image.Image.Project)
	_ = d.Set("organization_id", image.Image.Organization)

	return nil
}

func resourceScalewayInstanceImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(meta, d.Id())
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
	if d.HasChange("root_volume_id") {
		//TODO: I probably need to use an expandVolume type function here
		req.RootVolume = d.Get("root_volume_id").(*instance.VolumeSummary)
	}
	if d.HasChange("architecture") {
		req.Arch = instance.Arch(d.Get("architecture").(string))
	}
	if d.HasChange("default_bootscript_id") {
		bootscriptSpecs, err := getBootscriptSpecs(d.Get("default_bootscript_id").(string), instanceAPI, zone, ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		req.DefaultBootscript = bootscriptSpecs
	}
	if d.HasChange("additional_volume_ids") {
		//TODO: verify that this works
		snapResponses, err := getExtraVolumesSpecsFromSnapshots(d.Get("additional_volume_ids").([]interface{}), instanceAPI, ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		req.ExtraVolumes = expandInstanceImageExtraVolumes(snapResponses)
	}
	if d.HasChange("tags") {
		req.Tags = expandUpdatedStringsPtr(d.Get("tags"))
	}
	if d.HasChange("public") {
		req.Public = d.Get("public").(bool)
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

	return resourceScalewayInstanceImageRead(ctx, d, meta)
}

func resourceScalewayInstanceImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, id, err := instanceAPIWithZoneAndID(meta, d.Id())
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
		if !is404Error(err) {
			return diag.FromErr(err)
		}
	}

	_, err = waitForInstanceImage(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if !is404Error(err) {
			return diag.FromErr(err)
		}
	}

	return nil
}
