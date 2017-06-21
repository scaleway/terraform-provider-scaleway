package scaleway

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/nicolai86/scaleway-sdk/api"
)

func dataSourceScalewayImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayImageRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"name_filter": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"architecture": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Computed values.
			"organization": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func scalewayImageAttributes(d *schema.ResourceData, img *api.ScalewayImage) error {
	d.Set("architecture", img.Arch)
	d.Set("organization", img.Organization)
	d.Set("public", img.Public)
	d.Set("creation_date", img.CreationDate)
	d.Set("name", img.Name)
	d.SetId(img.Identifier)

	return nil
}

func dataSourceScalewayImageRead(d *schema.ResourceData, meta interface{}) error {
	scaleway := meta.(*Client).scaleway

	var nameMatch func(api.MarketImage) bool
	if name, ok := d.GetOk("name"); ok {
		nameMatch = func(img api.MarketImage) bool {
			return img.Name == name.(string)
		}
	} else if nameFilter, ok := d.GetOk("name_filter"); ok {
		exp := regexp.MustCompile(nameFilter.(string))
		nameMatch = func(img api.MarketImage) bool {
			return exp.MatchString(img.Name)
		}
	}

	imgs, err := scaleway.GetImages()
	if err != nil {
		return err
	}
	images := []api.MarketLocalImageDefinition{}
	for _, image := range *imgs {
		if !nameMatch(image) {
			continue
		}

		for _, v := range image.Versions {
			for _, l := range v.LocalImages {
				if l.Arch == d.Get("architecture").(string) && l.Zone == scaleway.Region {
					images = append(images, l)
				}
			}
		}
	}

	if len(images) > 1 {
		return fmt.Errorf("The query returned more than one result. Please refine your query.")
	}
	if len(images) == 0 {
		return fmt.Errorf("The query returned no result. Please refine your query.")
	}

	img, err := scaleway.GetImage(images[0].ID)
	if err != nil {
		return err
	}

	return scalewayImageAttributes(d, img)
}
