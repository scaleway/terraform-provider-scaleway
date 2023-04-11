package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayWebhostingOffer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayWebhostingOfferRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Exact name of the desired offer",
				ConflictsWith: []string{"offer_id"},
			},
			"offer_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of the desired offer",
				ConflictsWith: []string{"name"},
			},
			"billing_operation_path": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"product": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"option": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"email_accounts_quota": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"email_storage_quota": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"databases_quota": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"hosting_storage_quota": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"support_included": {
							Computed: true,
							Type:     schema.TypeBool,
						},
						"v_cpu": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"ram": {
							Computed: true,
							Type:     schema.TypeInt,
						},
					},
				},
			},
			"price": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": regionSchema(),
		},
	}
}

func dataSourceScalewayWebhostingOfferRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	webhostingAPI, region, err := webhostingAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := webhostingAPI.ListOffers(&webhosting.ListOffersRequest{
		Region: region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	if len(res.Offers) == 0 {
		return diag.FromErr(fmt.Errorf("no offer found in region %s", region))
	}

	var filteredOffer *webhosting.Offer
	for _, offer := range res.Offers {
		if offer.ID == d.Get("offer_id") || offer.Product.Name == d.Get("name") {
			filteredOffer = offer
		}
	}
	if filteredOffer == nil {
		return diag.FromErr(fmt.Errorf("no offer found with the name or id: %s%s in region %s", d.Get("name"), d.Get("offer_id"), region))
	}

	regionalID := datasourceNewRegionalizedID(filteredOffer.ID, region)
	d.SetId(regionalID)
	_ = d.Set("offer_id", regionalID)
	_ = d.Set("name", filteredOffer.Product.Name)
	_ = d.Set("region", region)
	_ = d.Set("billing_operation_path", filteredOffer.BillingOperationPath)
	_ = d.Set("product", flattenOfferProduct(filteredOffer.Product))
	_ = d.Set("price", flattenOfferPrice(filteredOffer.Price))

	return nil
}
