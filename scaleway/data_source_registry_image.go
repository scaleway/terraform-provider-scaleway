package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func dataSourceScalewayRegistryImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayRegistryImageRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of the registry image",
				ConflictsWith: []string{"image_id"},
			},
			"image_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The ID of the registry image",
				ConflictsWith: []string{"name"},
				ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
			},
			"namespace_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "The namespace ID of the registry image",
				ValidateFunc: verify.IsUUIDorUUIDWithLocality(),
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the registry image",
			},
			"visibility": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The visibility policy of the registry image",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The tags associated with the registry image",
			},
			"updated_at": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"region":          regional.Schema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func dataSourceScalewayRegistryImageRead(d *schema.ResourceData, m interface{}) error {
	api, region, err := registryAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	var image *registry.Image
	imageID, ok := d.GetOk("image_id")
	if !ok {
		var namespaceID *string
		if d.Get("namespace_id") != "" {
			namespaceID = types.ExpandStringPtr(locality.ExpandID(d.Get("namespace_id")))
		}
		imageName := d.Get("name").(string)
		res, err := api.ListImages(&registry.ListImagesRequest{
			Region:      region,
			Name:        types.ExpandStringPtr(imageName),
			NamespaceID: namespaceID,
			ProjectID:   types.ExpandStringPtr(d.Get("project_id")),
		})
		if err != nil {
			return err
		}
		foundImage, err := findExact(
			res.Images,
			func(s *registry.Image) bool { return s.Name == imageName },
			imageName,
		)
		if err != nil {
			return err
		}

		image = foundImage
	} else {
		res, err := api.GetImage(&registry.GetImageRequest{
			Region:  region,
			ImageID: locality.ExpandID(imageID),
		})
		if err != nil {
			return err
		}
		image = res
	}

	d.SetId(datasource.NewRegionalID(image.ID, region))
	_ = d.Set("image_id", image.ID)
	_ = d.Set("name", image.Name)
	_ = d.Set("namespace_id", image.NamespaceID)
	_ = d.Set("visibility", image.Visibility.String())
	_ = d.Set("size", int(image.Size))
	_ = d.Set("tags", image.Tags)
	_ = d.Set("updated_at", types.FlattenTime(image.UpdatedAt))

	return nil
}
