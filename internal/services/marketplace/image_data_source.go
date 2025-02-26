package marketplace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceMarketplaceImageRead,
		Schema: map[string]*schema.Schema{
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Exact label of the desired image",
			},
			"instance_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "DEV1-S",
				Description: "The instance commercial type of the desired image",
			},
			"image_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "instance_local", // Keep the old default as default to avoid a breaking change.
				Description: "The type of the desired image, instance_local or instance_sbs",
			},
			"zone": zonal.Schema(),
		},
	}
}

func DataSourceMarketplaceImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	marketplaceAPI, zone, err := marketplaceAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	image, err := marketplaceAPI.GetLocalImageByLabel(&marketplace.GetLocalImageByLabelRequest{
		ImageLabel:     d.Get("label").(string),
		CommercialType: d.Get("instance_type").(string),
		Zone:           zone,
		Type:           marketplace.LocalImageType(d.Get("image_type").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	zonedID := datasource.NewZonedID(image.ID, zone)
	d.SetId(zonedID)
	_ = d.Set("zone", zone)
	_ = d.Set("label", d.Get("label"))
	_ = d.Set("instance_type", d.Get("type"))
	_ = d.Set("image_type", image.Type)

	return nil
}
