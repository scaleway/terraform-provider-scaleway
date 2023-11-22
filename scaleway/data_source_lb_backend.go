package scaleway

import (
	"context"

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

	backID, ok := d.GetOk("backend_id")
	if !ok { // Get LB by name.
		backendName := d.Get("name").(string)
		res, err := api.ListBackends(&lbSDK.ZonedAPIListBackendsRequest{
			Zone: zone,
			Name: expandStringPtr(backendName),
			LBID: expandID(d.Get("lb_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundBackend, err := findExact(
			res.Backends,
			func(s *lbSDK.Backend) bool { return s.Name == backendName },
			backendName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		backID = foundBackend.ID
	}
	zonedID := datasourceNewZonedID(backID, zone)
	d.SetId(zonedID)
	err = d.Set("backend_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayLbBackendRead(ctx, d, meta)
}
