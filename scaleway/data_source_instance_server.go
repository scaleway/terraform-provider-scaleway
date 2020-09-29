package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
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
		Read: dataSourceScalewayInstanceServerRead,

		Schema: dsSchema,
	}
}

func dataSourceScalewayInstanceServerRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return err
	}

	serverID, ok := d.GetOk("server_id")
	if !ok {
		res, err := instanceApi.ListServers(&instance.ListServersRequest{
			Zone: zone,
			Name: String(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
		if len(res.Servers) == 0 {
			return fmt.Errorf("no server found with the name %s", d.Get("name"))
		}
		if len(res.Servers) > 1 {
			return fmt.Errorf("%d servers found with the same name %s", len(res.Servers), d.Get("name"))
		}
		serverID = res.Servers[0].ID
	}

	zonedID := datasourceNewZonedID(serverID, zone)
	d.SetId(zonedID)
	_ = d.Set("server_id", zonedID)
	return resourceScalewayInstanceServerRead(d, m)
}
