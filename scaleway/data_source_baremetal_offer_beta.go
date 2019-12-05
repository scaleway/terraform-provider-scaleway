package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	baremetal "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayBaremetalOfferBeta() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayBaremetalOfferBetaRead,

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
			"allow_disabled": {
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
						"ecc": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if error-correcting code is available on this memory",
						},
					},
				},
			},
			"price_per_sixty_minutes": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Price of the offer for the next 60 minutes (a server order at 11h32 will be payed until 12h32)",
			},
			"price_per_month": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Price of the offer per months",
			},
			"quota_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Quota name of this offer",
			},
			"stock": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Stock status for this offer",
			},
		},
	}
}

func dataSourceScalewayBaremetalOfferBetaRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	baremetalApi, fallBackZone, err := getBaremetalAPIWithZone(d, meta)
	if err != nil {
		return err
	}

	zone, offerID, _ := parseZonedID(datasourceNewZonedID(d.Get("offer_id"), fallBackZone))
	res, err := baremetalApi.ListOffers(&baremetal.ListOffersRequest{
		Zone: zone,
	}, scw.WithAllPages())
	if err != nil {
		return err
	}
	excludedOffers := []int(nil)
	for i, offer := range res.Offers {
		switch {
		case offer.Name == d.Get("name"), offer.ID == offerID:
			if !offer.Enable && !d.Get("allow_disabled").(bool) {
				return fmt.Errorf("offer %s (%s) found in zone %s but is disabled. Add allow_disabled=true in your terraform config to use it.", offer.Name, offer.Name, zone)
			}
		default:
			excludedOffers = append(excludedOffers, i)
		}
	}
	for _, excludedOffer := range excludedOffers {
		res.Offers = append(res.Offers[:excludedOffer], res.Offers[:excludedOffer+1]...)
	}

	if len(res.Offers) == 0 {
		return fmt.Errorf("no offer found with the name %s in zone %s", d.Get("name"), zone)
	}
	if len(res.Offers) > 1 {
		return fmt.Errorf("%d offers found with the same name %s in zone %s", len(res.Offers), d.Get("name"), zone)
	}

	offer := res.Offers[0]
	zonedID := datasourceNewZonedID(offer.ID, zone)
	d.SetId(zonedID)
	d.Set("offer_id", zonedID)
	d.Set("zone", zone)

	if err != nil {
		return err
	}

	d.Set("allow_disabled", !offer.Enable)
	d.Set("bandwidth", offer.Bandwidth)
	d.Set("commercial_range", offer.CommercialRange)
	d.Set("cpu", flattenCPUs(offer.CPU))
	d.Set("disk", flattenDisks(offer.Disk))
	d.Set("memory", flattenMemory(offer.Memory))
	d.Set("price_per_sixty_minutes", flattenMoney(offer.PricePerSixtyMinutes))
	d.Set("price_per_month", flattenMoney(offer.PricePerMonth))
	d.Set("quota_name", offer.QuotaName)
	d.Set("stock", offer.Stock.String())

	return nil
}
