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
)

//go:embed descriptions/partners_data_source.md
var partnersDataSourceDescription string

func DataSourcePartners() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourcePartnersRead,
		Description: partnersDataSourceDescription,
		Schema: map[string]*schema.Schema{
			"pop_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Filter for partners present in one of these PoPs",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"region": regional.Schema(),
			// Computed
			"partners": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of partners",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the partner",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the partner",
						},
						"contact_email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Contact email address",
						},
						"logo_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL of the partner's logo",
						},
						"portal_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL of the partner's portal",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation date",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last update date",
						},
					},
				},
			},
		},
	}
}

func DataSourcePartnersRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &interlink.ListPartnersRequest{
		Region: region,
	}

	if popIDs, ok := d.GetOk("pop_ids"); ok {
		req.PopIDs = locality.ExpandIDs(popIDs)
	}

	res, err := api.ListPartners(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(region.String())
	_ = d.Set("partners", flattenPartners(region, res.Partners))
	_ = d.Set("region", region.String())

	return nil
}
