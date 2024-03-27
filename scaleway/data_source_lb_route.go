package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceScalewayLbRoute() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceScalewayLbRoute().Schema)

	dsSchema["route_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "The ID of the route",
		ValidateFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayLbRouteRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayLbRouteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	return resourceScalewayLbRouteRead(ctx, d, m)
}
