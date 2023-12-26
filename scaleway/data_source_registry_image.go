package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
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
				ValidateFunc:  validationUUIDorUUIDWithLocality(),
			},
			"namespace_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "The namespace ID of the registry image",
				ValidateFunc: validationUUIDorUUIDWithLocality(),
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
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func dataSourceScalewayRegistryImageRead(d *schema.ResourceData, meta interface{}) error {
	api, region, err := registryAPIWithRegion(d, meta)
	if err != nil {
		return err
	}

	var image *registry.Image
	imageID, ok := d.GetOk("image_id")
	if !ok {
		var namespaceID *string
		if d.Get("namespace_id") != "" {
			namespaceID = expandStringPtr(expandID(d.Get("namespace_id")))
		}
		imageName := d.Get("name").(string)
		res, err := api.ListImages(&registry.ListImagesRequest{
			Region:      region,
			Name:        expandStringPtr(imageName),
			NamespaceID: namespaceID,
			ProjectID:   expandStringPtr(d.Get("project_id")),
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
			ImageID: expandID(imageID),
		})
		if err != nil {
			return err
		}
		image = res
	}

	d.SetId(datasourceNewRegionalID(image.ID, region))
	_ = d.Set("image_id", image.ID)
	_ = d.Set("name", image.Name)
	_ = d.Set("namespace_id", image.NamespaceID)
	_ = d.Set("visibility", image.Visibility.String())
	_ = d.Set("size", int(image.Size))
	_ = d.Set("tags", image.Tags)
	_ = d.Set("updated_at", flattenTime(image.UpdatedAt))

	return nil
}
