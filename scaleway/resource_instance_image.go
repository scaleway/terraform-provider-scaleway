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
				//Computed:    true, // TODO : maybe ? to set it to the default value
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

	//defaultBootscript, bootscriptExists := d.GetOk("default_bootscript_id")
	//if bootscriptExists {
	//	req.DefaultBootscript = expandStrings(defaultBootscript)[0]
	//}
	extraVolumesIds, volumesExist := d.GetOk("additional_volume_ids")
	if volumesExist {
		snapResponses, err := getExtraVolumesSpecsFromSnapshots(extraVolumesIds.([]interface{}), instanceAPI, ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		req.ExtraVolumes = expandInstanceImageExtraVolumes(snapResponses)
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
	_ = d.Set("default_bootscript_id", image.Image.DefaultBootscript)
	if _, extraVolumesExist := d.GetOk("additional_volume_ids"); extraVolumesExist == true {
		//additionalVolumeIDs := []string(nil)
		//for _, volume := range orderVolumes(image.Image.ExtraVolumes) {
		//	additionalVolumeIDs = append(additionalVolumeIDs, volume.ID)
		//}
		//_ = d.Set("additional_volume_ids", additionalVolumeIDs)
		_ = d.Set("additional_volume_ids", flattenInstanceImageExtraVolumes(image.Image.ExtraVolumes))
	}
	_ = d.Set("tags", image.Image.Tags)
	_ = d.Set("public", image.Image.Public)
	_ = d.Set("creation_date", image.Image.CreationDate.Format(time.RFC3339))
	_ = d.Set("modification_date", image.Image.ModificationDate.Format(time.RFC3339))
	_ = d.Set("from_server_id", image.Image.FromServer)
	_ = d.Set("state", image.Image.State)
	//_ = d.Set("location", image.Image.Lo)
	_ = d.Set("zone", image.Image.Zone)
	_ = d.Set("project_id", image.Image.Project)
	_ = d.Set("organization_id", image.Image.Organization)

	return nil
}

func resourceScalewayInstanceImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//instanceAPI, zone, id, err := instanceAPIWithZoneAndID(meta, d.Id())
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	////TODO: UpdateImage types are not implemented !!
	//req := &instance.UpdateImageRequest{
	//	ImageID: id,
	//	Zone:    zone,
	//	Name:    scw.StringPtr(d.Get("name").(string)),
	//	Tags:    scw.StringsPtr([]string{}),
	//}
	//
	//_, err = instanceAPI.UpdateImage(req, scw.WithContext(ctx))
	//if err != nil {
	//	return diag.FromErr(fmt.Errorf("couldn't update image: %s", err))
	//}

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
