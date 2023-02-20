package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScalewayLbRoute() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayLbRoute().Schema)

	dsSchema["route_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "The ID of the route",
		ValidateFunc: validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayLbRouteRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayLbRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	routeID, _ := d.GetOk("route_id")

	zonedID := datasourceNewZonedID(routeID, zone)
	d.SetId(zonedID)
	err = d.Set("route_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayLbRouteRead(ctx, d, meta)
}
