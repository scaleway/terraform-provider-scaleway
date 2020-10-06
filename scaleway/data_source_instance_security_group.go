package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstanceSecurityGroup() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstanceSecurityGroup().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "zone")

	dsSchema["name"].ConflictsWith = []string{"security_group_id"}
	dsSchema["security_group_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the security group",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		Read: dataSourceScalewayInstanceSecurityGroupRead,

		Schema: dsSchema,
	}
}

func dataSourceScalewayInstanceSecurityGroupRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return err
	}

	securityGroupID, ok := d.GetOk("security_group_id")
	if !ok {
		res, err := instanceAPI.ListSecurityGroups(&instance.ListSecurityGroupsRequest{
			Zone: zone,
			Name: expandStringPtr(d.Get("name")),
		}, scw.WithAllPages())
		if err != nil {
			return err
		}
		for _, sg := range res.SecurityGroups {
			if sg.Name == d.Get("name").(string) {
				if securityGroupID != "" {
					return fmt.Errorf("more than 1 security group found with the same name %s", d.Get("name"))
				}
				securityGroupID = sg.ID
			}
		}
		if securityGroupID == "" {
			return fmt.Errorf("no security group found with the name %s", d.Get("name"))
		}
	}

	zonedID := datasourceNewZonedID(securityGroupID, zone)
	d.SetId(zonedID)
	_ = d.Set("security_group_id", zonedID)
	return resourceScalewayInstanceSecurityGroupRead(d, m)
}
