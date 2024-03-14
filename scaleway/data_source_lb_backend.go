package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceScalewayLbBackend() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceScalewayLbBackend().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "lb_id")

	dsSchema["name"].ConflictsWith = []string{"backend_id"}
	dsSchema["backend_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the backend",
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayLbBackendRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayLbBackendRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	backID, ok := d.GetOk("backend_id")
	if !ok { // Get LB by name.
		backendName := d.Get("name").(string)
		res, err := api.ListBackends(&lbSDK.ZonedAPIListBackendsRequest{
			Zone: zone,
			Name: types.ExpandStringPtr(backendName),
			LBID: locality.ExpandID(d.Get("lb_id")),
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
	zonedID := datasource.NewZonedID(backID, zone)
	d.SetId(zonedID)
	err = d.Set("backend_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayLbBackendRead(ctx, d, m)
}
