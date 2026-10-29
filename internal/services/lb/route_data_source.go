package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceRoute() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceRoute().SchemaFunc())

	dsSchema["route_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		Description:      "The ID of the route",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceLbRouteRead,
		Schema:      dsSchema,
	}
}

func DataSourceLbRouteRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	routeID, _ := d.GetOk("route_id")

	zonedID := datasource.NewZonedID(routeID, zone)
	d.SetId(zonedID)

	route, err := api.GetRoute(&lbSDK.ZonedAPIGetRouteRequest{
		Zone:    zone,
		RouteID: locality.ExpandID(routeID.(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setRouteState(d, route, zone)
}
