package webhosting

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func DataSourceOffer() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		ReadContext:                       dataSourceOfferRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Exact name of the desired offer",
				ConflictsWith: []string{"offer_id"},
			},
			"control_panel": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Name of the control panel.(Cpanel or Plesk)",
				DiffSuppressFunc: dsf.IgnoreCase,
				ConflictsWith:    []string{"offer_id"},
			},
			"offer_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of the desired offer",
				ConflictsWith: []string{"name", "control_panel"},
			},
			"billing_operation_path": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"product": {
				Type:       schema.TypeList,
				Computed:   true,
				Deprecated: "The product field is deprecated. Please use the offer field instead.",
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
							Type:     schema.TypeInt,
							Computed: true,
						},
						"email_storage_quota": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"databases_quota": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"hosting_storage_quota": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"support_included": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"v_cpu": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ram": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"offer": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The offer details of the hosting",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"billing_operation_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"available": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"control_panel_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_of_life": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"quota_warning": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"price": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"billing_operation_path": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"min_value": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"current_value": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"max_value": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"quota_warning": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"price": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"price": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": regional.Schema(),
		},
	}
}

func dataSourceOfferRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newOfferAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.ListOffers(&webhosting.OfferAPIListOffersRequest{
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
		cp, _ := d.Get("control_panel").(string)
		if offer.ID == d.Get("offer_id") || (offer.Name == d.Get("name") && (cp == "" || strings.EqualFold(offer.ControlPanelName, cp))) {
			filteredOffer = offer

			break
		}
	}

	if filteredOffer == nil {
		return diag.FromErr(fmt.Errorf("no offer found with the name or id: %s%s in region %s", d.Get("name"), d.Get("offer_id"), region))
	}

	regionalID := datasource.NewRegionalID(filteredOffer.ID, region)
	d.SetId(regionalID)
	_ = d.Set("offer_id", regionalID)
	_ = d.Set("name", filteredOffer.Name)
	_ = d.Set("region", region)
	_ = d.Set("billing_operation_path", filteredOffer.BillingOperationPath)
	_ = d.Set("product", nil)
	_ = d.Set("offer", flattenOffer(filteredOffer))
	_ = d.Set("price", flattenOfferPrice(filteredOffer.Price))

	return nil
}
