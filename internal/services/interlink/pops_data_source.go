package interlink

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/pops_data_source.md
var popsDataSourceDescription string

func DataSourcePops() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourcePopsRead,
		Description: popsDataSourceDescription,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "PoP name to filter for",
			},
			"hosting_provider_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Hosting provider name to filter for",
			},
			"partner_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter for PoPs hosting an available shared connection from this partner",
			},
			"link_bandwidth_mbps": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Filter for PoPs with a shared connection allowing this bandwidth size",
			},
			"dedicated_available": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filter for PoPs with a dedicated connection available for self-hosted links",
			},
			"region": regional.Schema(),
			// Computed
			"pops": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of PoPs",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the PoP",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the PoP",
						},
						"hosting_provider_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the PoP's hosting provider",
						},
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Physical address of the PoP",
						},
						"city": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "City where the PoP is located",
						},
						"logo_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL of the PoP's logo",
						},
						"available_link_bandwidths_mbps": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Available bandwidth options in Mbps",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Human-readable display name",
						},
						"region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Region of the PoP",
						},
					},
				},
			},
		},
	}
}

func DataSourcePopsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &interlink.ListPopsRequest{
		Region:              region,
		Name:                types.ExpandStringPtr(d.Get("name")),
		HostingProviderName: types.ExpandStringPtr(d.Get("hosting_provider_name")),
	}

	if partnerID, ok := d.GetOk("partner_id"); ok {
		req.PartnerID = new(locality.ExpandID(partnerID.(string)))
	}

	if bandwidth, ok := d.GetOk("link_bandwidth_mbps"); ok {
		req.LinkBandwidthMbps = new(uint64(bandwidth.(int)))
	}

	if dedicated, ok := d.GetOk("dedicated_available"); ok {
		req.DedicatedAvailable = new(dedicated.(bool))
	}

	res, err := api.ListPops(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(region.String())
	_ = d.Set("pops", flattenPops(res.Pops))
	_ = d.Set("region", region.String())

	return nil
}
