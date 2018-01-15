package scaleway

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func dataSourceScalewayImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayImageRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "exact name of the desired image",
			},
			"name_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "partial name of the desired image to filter with",
			},
			"architecture": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "architecture of the desired image",
			},
			// Computed values.
			"organization": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "organization owning the bootscript",
			},
			"public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "indication if the bootscript is public",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "date when the image was created",
			},
		},
	}
}

func scalewayImageAttributes(d *schema.ResourceData, img *api.Image) error {
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
		exp, err := regexp.Compile(nameFilter.(string))
		if err != nil {
			return fmt.Errorf("invalid name_filter regular expression provided: %v", err)
		}
		nameMatch = func(img api.MarketImage) bool {
			return exp.MatchString(img.Name)
		}
	}

	var (
		imgs *[]api.MarketImage
		err  error
	)
	if err := retry(func() error {
		imgs, err = scaleway.GetImages()
		return err
	}); err != nil {
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

	var img *api.Image
	if err := retry(func() error {
		img, err = scaleway.GetImage(images[0].ID)
		return err
	}); err != nil {
		return err
	}

	return scalewayImageAttributes(d, img)
}
