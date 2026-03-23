package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceCockpitProducts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCockpitProductsRead,
		Schema: map[string]*schema.Schema{
			"region": regional.Schema(),
			"products": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of Cockpit products available for exported_products in scaleway_cockpit_exporter.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Product name to use in exported_products (e.g. cockpit, LB, object-storage).",
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Human-readable display name of the product.",
						},
						"family_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Product family name.",
						},
					},
				},
			},
			"names": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of product names for use in scaleway_cockpit_exporter.exported_products.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceCockpitProductsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := api.ListProducts(&cockpit.RegionalAPIListProductsRequest{
		Region:  region,
		OrderBy: cockpit.ListProductsRequestOrderByDisplayNameAsc,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	products := make([]map[string]any, 0, len(resp.ProductsList))
	names := make([]string, 0, len(resp.ProductsList))

	for _, p := range resp.ProductsList {
		products = append(products, map[string]any{
			"name":         p.Name,
			"display_name": p.DisplayName,
			"family_name":  p.FamilyName,
		})
		names = append(names, p.Name)
	}

	d.SetId(region.String())
	_ = d.Set("region", region.String())
	_ = d.Set("products", products)
	_ = d.Set("names", types.FlattenSliceString(names))

	return nil
}
