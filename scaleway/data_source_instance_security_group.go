package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func dataSourceScalewayInstanceSecurityGroup() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstanceSecurityGroup().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "organization_id", "zone")

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
	instanceApi, zone, err := getInstanceAPIWithZone(d, meta)
	if err != nil {
		return err
	}

	var securityGroup *instance.SecurityGroup
	securityGroupID, ok := d.GetOk("security_group_id")
	if ok {
		res, err := instanceApi.GetSecurityGroup(&instance.GetSecurityGroupRequest{
			Zone:            zone,
			SecurityGroupID: expandID(securityGroupID),
		})
		if err != nil {
			return err
		}
		securityGroup = res.SecurityGroup
	} else {
		res, err := instanceApi.ListSecurityGroups(&instance.ListSecurityGroupsRequest{
			Zone: zone,
			Name: String(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
		if len(res.SecurityGroups) == 0 {
			return fmt.Errorf("no security group found with the name %s", d.Get("name"))
		}
		if len(res.SecurityGroups) > 1 {
			return fmt.Errorf("%d security groups found with the same name %s", len(res.SecurityGroups), d.Get("name"))
		}
		securityGroup = res.SecurityGroups[0]
	}

	d.SetId(newZonedId(zone, securityGroup.ID))
	d.Set("security_group_id", newZonedId(zone, securityGroup.ID))
	return resourceScalewayInstanceSecurityGroupRead(d, m)
}
