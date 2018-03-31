package scaleway

import (
	"errors"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewaySecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewaySecurityGroupCreate,
		Read:   resourceScalewaySecurityGroupRead,
		Update: resourceScalewaySecurityGroupUpdate,
		Delete: resourceScalewaySecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the security group",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The description of the security group",
			},
			"enable_default_security": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
				Description: "Add default security group rules",
			},
		},
	}
}

func resourceScalewaySecurityGroupCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	req := api.NewSecurityGroup{
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		Organization:          scaleway.Organization,
		EnableDefaultSecurity: d.Get("enable_default_security").(bool),
	}

	group, err := scaleway.CreateSecurityGroup(req)
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error creating security group: %q\n", serr.APIMessage)
		}

		return err
	}

	d.SetId(group.ID)

	if d.Id() == "" {
		return errors.New("failed to find created security group")
	}

	return resourceScalewaySecurityGroupRead(d, m)
}

func resourceScalewaySecurityGroupRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway
	group, err := scaleway.GetSecurityGroup(d.Id())

	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error reading security group: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}

		return err
	}

	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("enable_default_security", group.EnableDefaultSecurity)

	return nil
}

func resourceScalewaySecurityGroupUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	var req = api.UpdateSecurityGroup{
		Organization: scaleway.Organization,
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
	}

	if _, err := scaleway.UpdateSecurityGroup(req, d.Id()); err != nil {
		log.Printf("[DEBUG] Error reading security group: %q\n", err)

		return err
	}

	return resourceScalewaySecurityGroupRead(d, m)
}

func resourceScalewaySecurityGroupDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	err := scaleway.DeleteSecurityGroup(d.Id())
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] error reading Security Group Rule: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}

		return err
	}

	d.SetId("")
	return nil
}
