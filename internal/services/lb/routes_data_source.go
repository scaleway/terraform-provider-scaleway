package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceRoutes() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceLbRoutesRead,
		SchemaFunc:  routesSchema,
	}
}

func routesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"frontend_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Routes with a frontend id like it are listed.",
		},
		"routes": {
			Type:        schema.TypeList,
			Description: "List of routes.",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Computed:    true,
						Description: "UUID of the route.",
						Type:        schema.TypeString,
					},
					"frontend_id": {
						Computed:    true,
						Description: "UUID of the frontend to use for this route.",
						Type:        schema.TypeString,
					},
					"backend_id": {
						Computed:    true,
						Description: "UUID of the backend to use for this route.",
						Type:        schema.TypeString,
					},
					"match_sni": {
						Computed:    true,
						Description: "Server Name Indication (SNI) value to match. Value to match in the Server Name Indication TLS extension (SNI) field from an incoming connection made via an SSL/TLS transport layer. This field should be set for routes on TCP Load Balancers. ",
						Type:        schema.TypeString,
					},
					"match_host_header": {
						Computed:    true,
						Description: "HTTP host header to match. Value to match in the HTTP Host request header from an incoming request. This field should be set for routes on HTTP Load Balancers.",
						Type:        schema.TypeString,
					},
					"match_subdomains": {
						Computed:    true,
						Description: "If true, all subdomains of this route will be matched for this route.",
						Type:        schema.TypeBool,
					},
					"created_at": {
						Computed:    true,
						Description: "Date on which the route was created (RFC 3339 format)",
						Type:        schema.TypeString,
					},
					"update_at": {
						Computed:    true,
						Description: "Date at which the route was last updated (RFC 3339 format)",
						Type:        schema.TypeString,
					},
				},
			},
		},
		"zone":            zonal.Schema(),
		"organization_id": account.OrganizationIDSchema(),
		"project_id":      account.ProjectIDSchema(),
	}
}

func DataSourceLbRoutesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_, frontID, err := zonal.ParseID(d.Get("frontend_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := lbAPI.ListRoutes(&lb.ZonedAPIListRoutesRequest{
		Zone:       zone,
		FrontendID: types.ExpandStringPtr(frontID),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	routes := []any(nil)

	for _, route := range res.Routes {
		rawRoute := make(map[string]any)
		rawRoute["id"] = zonal.NewID(zone, route.ID).String()
		rawRoute["frontend_id"] = route.FrontendID
		rawRoute["backend_id"] = route.BackendID
		rawRoute["created_at"] = types.FlattenTime(route.CreatedAt)
		rawRoute["update_at"] = types.FlattenTime(route.UpdatedAt)
		rawRoute["match_sni"] = types.FlattenStringPtr(route.Match.Sni)
		rawRoute["match_host_header"] = types.FlattenStringPtr(route.Match.HostHeader)
		rawRoute["match_subdomains"] = route.Match.MatchSubdomains

		routes = append(routes, rawRoute)
	}

	d.SetId(zone.String())
	_ = d.Set("routes", routes)

	return nil
}
