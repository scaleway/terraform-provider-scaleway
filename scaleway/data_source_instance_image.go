package scaleway

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstanceImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayInstanceImageRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Exact name of the desired image",
				ConflictsWith: []string{"image_id"},
			},
			"image_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of the desired image",
				ConflictsWith: []string{"name", "architecture"},
			},
			"architecture": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       instance.ArchX86_64.String(),
				Description:   "Architecture of the desired image",
				ConflictsWith: []string{"image_id"},
			},
			"latest": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       true,
				Description:   "Select most recent image if multiple match",
				ConflictsWith: []string{"image_id"},
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),

			"public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indication if the image is public",
			},
			"default_bootscript_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the bootscript associated with this image",
			},
			"root_volume_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the root volume associated with this image",
			},
			"additional_volume_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The additional volume IDs attached to the image",
			},
			"from_server_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the server the image is originated from",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date when the image was created",
			},
			"modification_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date when the image was updated",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "State of the image",
			},
		},
	}
}

func dataSourceScalewayInstanceImageRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi, zone, err := getInstanceAPIWithZone(d, meta)
	if err != nil {
		return err
	}

	imageID, ok := d.GetOk("image_id")
	if !ok { // Get instance by name, zone, and arch.
		res, err := instanceApi.ListImages(&instance.ListImagesRequest{
			Zone: zone,
			Name: expandStringPtr(d.Get("name")),
			Arch: expandStringPtr(d.Get("architecture")),
		}, scw.WithAllPages())
		if err != nil {
			return err
		}
		if len(res.Images) == 0 {
			return fmt.Errorf("no image found with the name %s and architecture %s in zone %s", d.Get("name"), d.Get("architecture"), zone)
		}
		if len(res.Images) > 1 && !d.Get("latest").(bool) {
			return fmt.Errorf("%d images found with the same name %s and architecture %s in zone %s", len(res.Images), d.Get("name"), d.Get("architecture"), zone)
		}
		sort.Slice(res.Images, func(i, j int) bool {
			return res.Images[i].ModificationDate.After(res.Images[j].ModificationDate)
		})
		imageID = res.Images[0].ID
	}

	zonedID := datasourceNewZonedID(imageID, zone)
	zone, imageID, _ = parseZonedID(zonedID)

	d.SetId(zonedID)
	d.Set("image_id", zonedID)
	d.Set("zone", zone)

	resp, err := instanceApi.GetImage(&instance.GetImageRequest{
		Zone:    zone,
		ImageID: imageID.(string),
	})
	if err != nil {
		return err
	}

	d.Set("organization_id", resp.Image.Organization)
	d.Set("architecture", resp.Image.Arch)
	d.Set("name", resp.Image.Name)

	d.Set("creation_date", flattenTime(&resp.Image.CreationDate))
	d.Set("modification_date", flattenTime(&resp.Image.ModificationDate))
	d.Set("public", resp.Image.Public)
	d.Set("from_server_id", resp.Image.FromServer)
	d.Set("state", resp.Image.State.String())

	if resp.Image.DefaultBootscript != nil {
		d.Set("default_bootscript_id", resp.Image.DefaultBootscript.ID)
	} else {
		d.Set("default_bootscript_id", "")
	}

	if resp.Image.RootVolume != nil {
		d.Set("root_volume_id", resp.Image.RootVolume.ID)
	} else {
		d.Set("root_volume_id", "")
	}

	additionalVolumeIDs := []string(nil)
	for _, volume := range orderVolumes(resp.Image.ExtraVolumes) {
		additionalVolumeIDs = append(additionalVolumeIDs, volume.ID)
	}
	d.Set("additional_volume_ids", additionalVolumeIDs)

	return nil
}
