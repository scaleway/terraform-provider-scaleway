package instance

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceInstanceImageRead,

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
			"zone":            zonal.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),

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

func DataSourceInstanceImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	imageID, ok := d.GetOk("image_id")
	if !ok { // Get instance by name, zone, and arch.
		res, err := instanceAPI.ListImages(&instance.ListImagesRequest{
			Zone:    zone,
			Name:    types.ExpandStringPtr(d.Get("name")),
			Arch:    types.ExpandStringPtr(d.Get("architecture")),
			Project: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		var matchingImages []*instance.Image

		for _, image := range res.Images {
			if image.Name == d.Get("name").(string) {
				matchingImages = append(matchingImages, image)
			}
		}

		if len(matchingImages) == 0 {
			return diag.FromErr(fmt.Errorf("no image found with the name %s and architecture %s in zone %s", d.Get("name"), d.Get("architecture"), zone))
		}

		if len(matchingImages) > 1 && !d.Get("latest").(bool) {
			return diag.FromErr(fmt.Errorf("%d images found with the same name %s and architecture %s in zone %s", len(matchingImages), d.Get("name"), d.Get("architecture"), zone))
		}

		sort.Slice(matchingImages, func(i, j int) bool {
			return matchingImages[i].ModificationDate.After(*matchingImages[j].ModificationDate)
		})

		for _, image := range matchingImages {
			if image.Name == d.Get("name").(string) {
				imageID = image.ID

				break
			}
		}
	}

	zonedID := datasource.NewZonedID(imageID, zone)
	zone, imageID, _ = zonal.ParseID(zonedID)

	d.SetId(zonedID)
	_ = d.Set("image_id", zonedID)
	_ = d.Set("zone", zone)

	resp, err := instanceAPI.GetImage(&instance.GetImageRequest{
		Zone:    zone,
		ImageID: imageID.(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("organization_id", resp.Image.Organization)
	_ = d.Set("project_id", resp.Image.Project)
	_ = d.Set("architecture", resp.Image.Arch)
	_ = d.Set("name", resp.Image.Name)

	_ = d.Set("creation_date", types.FlattenTime(resp.Image.CreationDate))
	_ = d.Set("modification_date", types.FlattenTime(resp.Image.ModificationDate))
	_ = d.Set("public", resp.Image.Public)
	_ = d.Set("from_server_id", resp.Image.FromServer)
	_ = d.Set("state", resp.Image.State.String())

	if resp.Image.DefaultBootscript != nil {
		_ = d.Set("default_bootscript_id", resp.Image.DefaultBootscript.ID)
	} else {
		_ = d.Set("default_bootscript_id", "")
	}

	if resp.Image.RootVolume != nil {
		_ = d.Set("root_volume_id", resp.Image.RootVolume.ID)
	} else {
		_ = d.Set("root_volume_id", "")
	}

	additionalVolumeIDs := []string(nil)
	for _, volume := range orderVolumes(resp.Image.ExtraVolumes) {
		additionalVolumeIDs = append(additionalVolumeIDs, volume.ID)
	}

	_ = d.Set("additional_volume_ids", additionalVolumeIDs)

	return nil
}
