package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func dataSourceScalewayBaremetalServer() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(resourceScalewayBaremetalServer().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone", "project_id")

	dsSchema["name"].ConflictsWith = []string{"server_id"}
	dsSchema["server_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the server",
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayBaremetalServerRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayBaremetalServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := baremetalAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	serverID, ok := d.GetOk("server_id")
	if !ok { // Get server by zone and name.
		serverName := d.Get("name").(string)
		res, err := api.ListServers(&baremetal.ListServersRequest{
			Zone:      zone,
			Name:      scw.StringPtr(serverName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundServer, err := findExact(
			res.Servers,
			func(s *baremetal.Server) bool { return s.Name == serverName },
			serverName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		serverID = foundServer.ID
	}

	zoneID := datasource.NewZonedID(serverID, zone)
	d.SetId(zoneID)
	err = d.Set("server_id", zoneID)
	if err != nil {
		return diag.FromErr(err)
	}
	diags := resourceScalewayBaremetalServerRead(ctx, d, m)
	if diags != nil {
		return diags
	}
	if d.Id() == "" {
		return diag.Errorf("baremetal server (%s) not found", zoneID)
	}
	return nil
}
