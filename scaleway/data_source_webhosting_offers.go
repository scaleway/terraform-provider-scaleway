package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayWebhostingOffers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayWebhostingOffersRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Exact name of the desired offer",
				ConflictsWith: []string{"offer_id", "only_options", "without_options"},
			},
			"offer_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of the desired offer",
				ConflictsWith: []string{"name", "only_options", "without_options"},
			},
			"only_options": {
				Type:          schema.TypeBool,
				Optional:      true,
				Description:   "Select only offers, no options",
				ConflictsWith: []string{"name", "offer_id", "without_options"},
			},
			"without_options": {
				Type:          schema.TypeBool,
				Optional:      true,
				Description:   "Select only options",
				ConflictsWith: []string{"name", "offer_id", "only_options"},
			},
			"offers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed: true,
							Type:     schema.TypeString,
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
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"currency_code": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"units": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"nanos": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"region": regionSchema(),
		},
	}
}

func dataSourceScalewayWebhostingOffersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	webhostingAPI, region, err := webhostingAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &webhosting.ListOffersRequest{
		Region: region,
	}

	if onlyOptionsRaw, ok := d.GetOk("only_options"); ok {
		req.OnlyOptions = onlyOptionsRaw.(bool)
	}

	if withoutOptionsRaw, ok := d.GetOk("without_options"); ok {
		req.WithoutOptions = withoutOptionsRaw.(bool)
	}

	res, err := webhostingAPI.ListOffers(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var offers []interface{}
	for _, offer := range res.Offers {
		rawOffer := make(map[string]interface{})
		if offer.ID == d.Get("offer_id") || offer.Product.Name == d.Get("name") {
			rawOffer["id"] = newRegionalIDString(region, offer.ID)
			rawOffer["billing_operation_path"] = offer.BillingOperationPath
			rawOffer["product"] = flattenOfferProduct(offer.Product)
			rawOffer["price"] = flattenOfferPrice(offer.Price)

			offers = append(offers, rawOffer)
		}
	}

	if len(offers) == 0 {
		return diag.FromErr(fmt.Errorf("no offer found with the name or id in region %s", region))
	}

	d.SetId(region.String())
	_ = d.Set("offers", offers)

	return nil
}
