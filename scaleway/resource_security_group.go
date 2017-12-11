package scaleway

import (
	"fmt"
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
		},
	}
}

func resourceScalewaySecurityGroupCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	req := api.NewSecurityGroup{
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Organization: scaleway.Organization,
	}

	if err := retry(func() error {
		return scaleway.PostSecurityGroup(req)
	}); err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error creating security group: %q\n", serr.APIMessage)
		}

		return err
	}

	var (
		resp *api.GetSecurityGroups
		err  error
	)
	if err = retry(func() error {
		resp, err = scaleway.GetSecurityGroups()
		return err
	}); err != nil {
		return err
	}

	for _, group := range resp.SecurityGroups {
		if group.Name == req.Name {
			d.SetId(group.ID)
			break
		}
	}

	if d.Id() == "" {
		return fmt.Errorf("Failed to find created security group.")
	}

	return resourceScalewaySecurityGroupRead(d, m)
}

func resourceScalewaySecurityGroupRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway
	var (
		resp *api.GetSecurityGroup
		err  error
	)
	if err = retry(func() error {
		resp, err = scaleway.GetASecurityGroup(d.Id())
		return err
	}); err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error reading security group: %q\n", serr.APIMessage)

			if serr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}

		return err
	}

	d.Set("name", resp.SecurityGroups.Name)
	d.Set("description", resp.SecurityGroups.Description)

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

	if err := retry(func() error {
		return scaleway.PutSecurityGroup(req, d.Id())
	}); err != nil {
		log.Printf("[DEBUG] Error reading security group: %q\n", err)

		return err
	}

	return resourceScalewaySecurityGroupRead(d, m)
}

func resourceScalewaySecurityGroupDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	if err := retry(func() error {
		return scaleway.DeleteSecurityGroup(d.Id())
	}); err != nil {
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
