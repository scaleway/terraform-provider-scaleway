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
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return err
	}

	serverID, ok := d.GetOk("server_id")
	if !ok {
		res, err := instanceAPI.ListServers(&instance.ListServersRequest{
			Zone: zone,
			Name: expandStringPtr(d.Get("name")),
		})
		if err != nil {
			return err
		}
		for _, instance := range res.Servers {
			if instance.Name == d.Get("name").(string) {
				if serverID != "" {
					return fmt.Errorf("more than 1 server found with the same name %s", d.Get("name"))
				}
				serverID = instance.ID
			}
		}
		if serverID == "" {
			return fmt.Errorf("no server found with the name %s", d.Get("name"))
		}
	}

	zonedID := datasourceNewZonedID(serverID, zone)
	d.SetId(zonedID)
	_ = d.Set("server_id", zonedID)
	return resourceScalewayInstanceServerRead(d, m)
}
