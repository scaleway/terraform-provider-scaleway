package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v1"
)

func dataSourceScalewayMarketplaceImageBeta() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayMarketplaceImageReadBeta,

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
			"zone": zoneSchema(),
		},
	}
}

func dataSourceScalewayMarketplaceImageReadBeta(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	marketplaceAPI := marketplace.NewAPI(meta.scwClient)
	zone, err := getZone(d, meta)
	if err != nil {
		return err
	}

	imageID, err := marketplaceAPI.GetLocalImageIDByLabel(&marketplace.GetLocalImageIDByLabelRequest{
		ImageLabel:     d.Get("label").(string),
		CommercialType: d.Get("instance_type").(string),
		Zone:           zone,
	})
	if err != nil {
		return err
	}

	zonedID := datasourceNewZonedID(imageID, zone)
	d.SetId(zonedID)
	d.Set("zone", zone)
	d.Set("label", d.Get("label"))
	d.Set("instance_type", d.Get("type"))

	return nil
}
