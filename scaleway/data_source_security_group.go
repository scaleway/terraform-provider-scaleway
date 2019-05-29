package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/nicolai86/scaleway-sdk"
)

func dataSourceScalewaySecurityGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScalewaySecurityGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the security group",
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_default_security": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}
func dataSourceScalewaySecurityGroupRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Meta).deprecatedClient

	groups, err := client.GetSecurityGroups()
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error reading Security Groups: %q\n", serr.APIMessage)
		}

		return err
	}

	name := d.Get("name").(string)

	var group *api.SecurityGroup
	for _, v := range groups {
		if v.Name == name {
			group = &v
			break
		}
	}

	if group == nil {
		return fmt.Errorf("Security Group with name %q was not found!", name)
	}

	d.SetId(group.ID)

	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("enable_default_security", group.EnableDefaultSecurity)

	return nil
}
