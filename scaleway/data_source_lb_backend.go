package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayLbBackend() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayLbBackend().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "lb_id")

	dsSchema["name"].ConflictsWith = []string{"backend_id"}
	dsSchema["backend_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the backend",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayLbBackendRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayLbBackendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	crtID, ok := d.GetOk("backend_id")
	if !ok { // Get LB by name.
		res, err := api.ListBackends(&lbSDK.ZonedAPIListBackendsRequest{
			Zone: zone,
			Name: expandStringPtr(d.Get("name")),
			LBID: expandID(d.Get("lb_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Backends) == 0 {
			return diag.FromErr(fmt.Errorf("no backends found with the name %s", d.Get("name")))
		}
		if len(res.Backends) > 1 {
			return diag.FromErr(fmt.Errorf("%d backend found with the same name %s", len(res.Backends), d.Get("name")))
		}
		crtID = res.Backends[0].ID
	}
	zonedID := datasourceNewZonedID(crtID, zone)
	d.SetId(zonedID)
	err = d.Set("backend_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayLbBackendRead(ctx, d, meta)
}
