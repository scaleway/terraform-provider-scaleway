package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayLbFrontend() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayLbFrontend().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "lb_id")

	dsSchema["name"].ConflictsWith = []string{"frontend_id"}
	dsSchema["frontend_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the frontend",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayLbFrontendRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayLbFrontendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	crtID, ok := d.GetOk("frontend_id")
	if !ok { // Get LB by name.
		res, err := api.ListFrontends(&lbSDK.ZonedAPIListFrontendsRequest{
			Zone: zone,
			Name: expandStringPtr(d.Get("name")),
			LBID: expandID(d.Get("lb_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Frontends) == 0 {
			return diag.FromErr(fmt.Errorf("no frontends found with the name %s", d.Get("name")))
		}
		if len(res.Frontends) > 1 {
			return diag.FromErr(fmt.Errorf("%d frontend found with the same name %s", len(res.Frontends), d.Get("name")))
		}
		crtID = res.Frontends[0].ID
	}
	zonedID := datasourceNewZonedID(crtID, zone)
	d.SetId(zonedID)
	err = d.Set("frontend_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayLbFrontendRead(ctx, d, meta)
}
