package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayRegistryImageBeta() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayRegistryImageBetaRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of the Registry Image",
				ConflictsWith: []string{"image_id"},
			},
			"image_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The ID of the Registry Image",
				ConflictsWith: []string{"name"},
				ValidateFunc:  validationUUIDorUUIDWithLocality(),
			},
			"namespace_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "The namespace ID of the Registry Image",
				ValidateFunc: validationUUIDorUUIDWithLocality(),
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the Registry Image",
			},
			"visibility": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Registry Image visibility policy",
			},
		},
	}
}

func dataSourceScalewayRegistryImageBetaRead(d *schema.ResourceData, m interface{}) error {
	api, region, err := registryNamespaceAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	var image *registry.Image
	imageID, ok := d.GetOk("image_id")
	if !ok {
		var namespaceID *string
		if d.Get("namespace_id") != "" {
			namespaceID = scw.StringPtr(expandID(d.Get("namespace_id")))
		}
		res, err := api.ListImages(&registry.ListImagesRequest{
			Region:      region,
			Name:        String(d.Get("name").(string)),
			NamespaceID: namespaceID,
		})
		if err != nil {
			return err
		}
		if len(res.Images) == 0 {
			return fmt.Errorf("no images found with the name %s", d.Get("name"))
		}
		if len(res.Images) > 1 {
			return fmt.Errorf("%d images found with the same name %s", len(res.Images), d.Get("name"))
		}
		image = res.Images[0]
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
	_ = d.Set("size", image.Size)

	return nil
}
