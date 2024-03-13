package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func dataSourceScalewayMarketplaceImage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayMarketplaceImageRead,
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
			"zone": zonal.Schema(),
		},
	}
}

func dataSourceScalewayMarketplaceImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	marketplaceAPI, zone, err := marketplaceAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	image, err := marketplaceAPI.GetLocalImageByLabel(&marketplace.GetLocalImageByLabelRequest{
		ImageLabel:     d.Get("label").(string),
		CommercialType: d.Get("instance_type").(string),
		Zone:           zone,
		Type:           marketplace.LocalImageTypeInstanceLocal,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	zonedID := datasource.NewZonedID(image.ID, zone)
	d.SetId(zonedID)
	_ = d.Set("zone", zone)
	_ = d.Set("label", d.Get("label"))
	_ = d.Set("instance_type", d.Get("type"))

	return nil
}
