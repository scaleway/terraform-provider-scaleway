package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstanceServer() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstanceServer().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "zone")

	dsSchema["name"].ConflictsWith = []string{"server_id"}
	dsSchema["server_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the server",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayInstanceServerRead,

		Schema: dsSchema,
	}
}

func dataSourceScalewayInstanceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(*Meta)
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	serverID, ok := d.GetOk("server_id")
	if !ok {
		res, err := instanceAPI.ListServers(&instance.ListServersRequest{
			Zone:    zone,
			Name:    expandStringPtr(d.Get("name")),
			Project: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		for _, server := range res.Servers {
			if server.Name == d.Get("name").(string) {
				if serverID != "" {
					return diag.FromErr(fmt.Errorf("more than 1 server found with the same name %s", d.Get("name")))
				}
				serverID = server.ID
			}
		}
		if serverID == "" {
			return diag.FromErr(fmt.Errorf("no server found with the name %s", d.Get("name")))
		}
	}

	zonedID := datasourceNewZonedID(serverID, zone)
	d.SetId(zonedID)
	_ = d.Set("server_id", zonedID)
	return resourceScalewayInstanceServerRead(ctx, d, m)
}
