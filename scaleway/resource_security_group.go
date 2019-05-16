package scaleway

import (
	"errors"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	api "github.com/nicolai86/scaleway-sdk"
)

var supportedDefaultTrafficPolicies = []string{"accept", "drop", ""}

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
			"stateful": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Mark security group as stateful",
			},
			"inbound_default_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "accept",
				Description:  "Default inbound traffic policy for this security group",
				ValidateFunc: validation.StringInSlice(supportedDefaultTrafficPolicies, true),
			},
			"outbound_default_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "accept",
				Description:  "Default outbound traffic policy for this security group",
				ValidateFunc: validation.StringInSlice(supportedDefaultTrafficPolicies, true),
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
	scaleway := m.(*Meta).deprecatedClient

	req := api.NewSecurityGroup{
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		Organization:          scaleway.Organization,
		EnableDefaultSecurity: d.Get("enable_default_security").(bool),
		Stateful:              d.Get("stateful").(bool),
		InboundDefaultPolicy:  d.Get("inbound_default_policy").(string),
		OutboundDefaultPolicy: d.Get("outbound_default_policy").(string),
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
	scaleway := m.(*Meta).deprecatedClient
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
	d.Set("stateful", group.Stateful)
	d.Set("inbound_default_policy", group.InboundDefaultPolicy)
	d.Set("outbound_default_policy", group.OutboundDefaultPolicy)

	return nil
}

func resourceScalewaySecurityGroupUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	var req = api.UpdateSecurityGroup{
		Organization:          scaleway.Organization,
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		InboundDefaultPolicy:  d.Get("inbound_default_policy").(string),
		OutboundDefaultPolicy: d.Get("outbound_default_policy").(string),
	}

	if _, err := scaleway.UpdateSecurityGroup(req, d.Id()); err != nil {
		log.Printf("[DEBUG] Error reading security group: %q\n", err)

		return err
	}

	return resourceScalewaySecurityGroupRead(d, m)
}

func resourceScalewaySecurityGroupDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

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
