package scaleway

import (
	"context"
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
			// TODO: use all of these values at least once
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
			// "extra_volumes": {
			//	Type:        schema.TypeMap,
			//	Optional:    true,
			//	Description: "Additional volumes attached to the image",
			//	Elem: &schema.Schema{
			//		Type: schema.TypeMap, // TODO: not sure about that, maybe i must list all nested attributes here, or maybe i can call the instance volume schema
			//		Elem: &schema.Schema{
			//			Type: schema.TypeString,
			//		},
			//	},
			// },
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Project ID of the image",
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
				Computed:    true, // TODO : maybe ? to set it to the default value
				Description: "If true, the image will be public",
			},
			// Computed
			// "image_id": { // TODO: maybe we only need that in datasource
			//	Type:        schema.TypeString,
			//	Optional:    true,
			//	Description: "ID of the image",
			// },
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
				Description: "??", // TODO: find a proper description of this attribute
			},
			"organization": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "??", // TODO: find a proper description of this attribute
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The state of the image [ available | creating | error ]",
			},
			"location": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "??", // TODO: find a proper description of this attribute
			},
			// Common
			"zone":            zoneSchema(),
			"project_id":      projectIDSchema(),      // TODO: do we need that ?
			"organization_id": organizationIDSchema(), // TODO: do we need that ?
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
		req.DefaultBootscript = expandStrings(defaultBootscript)[0]
	}
	//extraVolumesIds, volumesExist := d.GetOk("additional_volumes_ids")
	//if volumesExist {
	//	req.ExtraVolumes = expand(extraVolumesIds)
	//}
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
	_ = d.Set("creation_date", image.Image.CreationDate.Format(time.RFC3339))
	_ = d.Set("tags", image.Image.Tags)
	_ = d.Set("default_bootscript_id", image.Image.DefaultBootscript)

	return nil
}

func resourceScalewayInstanceImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
