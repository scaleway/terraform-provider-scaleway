package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func dataSourceScalewayInstanceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewayInstanceSecurityGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the security group",
			},
			"security_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "The ID of the security group",
				ValidateFunc: validationUUIDorUUIDWithLocality(),
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
		},
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
	d.Set("name", securityGroup.Name)
	d.Set("security_group_id", securityGroup.ID)
	d.Set("zone", zone)
	d.Set("organization_id", securityGroup.Organization)

	return nil
}
