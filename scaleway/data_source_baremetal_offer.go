package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayBaremetalOffer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayBaremetalOfferRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Exact name of the desired offer",
				ConflictsWith: []string{"offer_id"},
			},
			"subscription_period": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					baremetal.OfferSubscriptionPeriodUnknownSubscriptionPeriod.String(),
					baremetal.OfferSubscriptionPeriodHourly.String(),
					baremetal.OfferSubscriptionPeriodMonthly.String(),
				}, false),
				Description:   "Period of subscription the desired offer",
				ConflictsWith: []string{"offer_id"},
			},
			"offer_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of the desired offer",
				ConflictsWith: []string{"name"},
			},
			"include_disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Include disabled offers",
			},
			"zone": zoneSchema(),

			"bandwidth": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Available Bandwidth with the offer",
			},
			"commercial_range": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Commercial range of the offer",
			},
			"cpu": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "CPU specifications of the offer",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "CPU name",
						},
						"core_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of cores",
						},
						"frequency": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Frequency of the CPU",
						},
						"thread_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of threads",
						},
					},
				},
			},
			"disk": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Disk specifications of the offer",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of disk",
						},
						"capacity": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Capacity of the disk in byte",
						},
					},
				},
			},
			"memory": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Memory specifications of the offer",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of memory",
						},
						"capacity": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Capacity of the memory in byte",
						},
						"frequency": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Frequency of the memory",
						},
						"is_ecc": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if error-correcting code is available on this memory",
						},
					},
				},
			},
			"stock": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Stock status for this offer",
			},
		},
	}
}

func dataSourceScalewayBaremetalOfferRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	baremetalAPI, fallBackZone, err := baremetalAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, offerID, _ := parseZonedID(datasourceNewZonedID(d.Get("offer_id"), fallBackZone))

	var offer *baremetal.Offer

	if offerID != "" {
		// offer, err = baremetalAPI.GetOffer(&baremetal.GetOfferRequest{
		// 	OfferID: offerID,
		// 	Zone:    zone,
		// })
		// if err != nil {
		// 	if is404Error(err) {
		// 		return diag.Errorf("offer %s not found", offerID)
		// 	}

		// 	return diag.FromErr(err)
		// }
		//
		// Temporary fix because GetOffer doesn't fetch monthly subscription offers
		offer, err = baremetalFindOfferById(ctx, baremetalAPI, zone, offerID)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		listOffersRequest := &baremetal.ListOffersRequest{
			Zone: zone,
		}
		if subscriptionPeriod, ok := d.GetOk("subscription_period"); ok {
			listOffersRequest.SubscriptionPeriod = baremetal.OfferSubscriptionPeriod(subscriptionPeriod.(string))
		}

		res, err := baremetalAPI.ListOffers(listOffersRequest, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		matches := []*baremetal.Offer{}
		for _, offer := range res.Offers {
			if offer.Name == d.Get("name") {
				if !offer.Enable && !d.Get("include_disabled").(bool) {
					return diag.FromErr(fmt.Errorf("%s offer %s (%s) found in zone %s but is disabled. Add allow_disabled=true in your terraform config to use it", offer.SubscriptionPeriod, offer.Name, offer.ID, zone))
				}

				matches = append(matches, offer)
			}
		}

		if len(matches) == 0 {
			if subscriptionPeriod, ok := d.GetOk("subscription_period"); ok {
				return diag.FromErr(fmt.Errorf("no offer found with the name %s and %s subscription period in zone %s", d.Get("name"), subscriptionPeriod, zone))
			}

			return diag.FromErr(fmt.Errorf("no offer found with the name %s in zone %s", d.Get("name"), zone))
		}

		if len(matches) > 1 {
			if subscriptionPeriod, ok := d.GetOk("subscription_period"); ok {
				return diag.FromErr(fmt.Errorf("%d offers found with the same name %s and %s subscription period in zone %s", len(matches), d.Get("name"), subscriptionPeriod, zone))
			}

			return diag.FromErr(fmt.Errorf("%d offers found with the same name %s in zone %s", len(matches), d.Get("name"), zone))
		}

		offer = matches[0]
	}

	zonedID := datasourceNewZonedID(offer.ID, zone)
	d.SetId(zonedID)
	_ = d.Set("offer_id", zonedID)
	_ = d.Set("zone", zone)
	_ = d.Set("name", offer.Name)
	_ = d.Set("subscription_period", offer.SubscriptionPeriod)
	_ = d.Set("include_disabled", !offer.Enable)
	_ = d.Set("bandwidth", int(offer.Bandwidth))
	_ = d.Set("commercial_range", offer.CommercialRange)
	_ = d.Set("cpu", flattenBaremetalCPUs(offer.CPUs))
	_ = d.Set("disk", flattenBaremetalDisks(offer.Disks))
	_ = d.Set("memory", flattenBaremetalMemory(offer.Memories))
	_ = d.Set("stock", offer.Stock.String())

	return nil
}
