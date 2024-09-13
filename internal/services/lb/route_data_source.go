package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceRoute() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceRoute().Schema)

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

func DataSourceLbRouteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	routeID, _ := d.GetOk("route_id")

	zonedID := datasource.NewZonedID(routeID, zone)
	d.SetId(zonedID)
	err = d.Set("route_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceLbRouteRead(ctx, d, m)
}
