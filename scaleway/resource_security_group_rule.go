package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewaySecurityGroupRule() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: `This resource is deprecated and will be removed in the next major version.
 Please use scaleway_instance_security_group_rule instead.`,

		Create: resourceScalewaySecurityGroupRuleCreate,
		Read:   resourceScalewaySecurityGroupRuleRead,
		Update: resourceScalewaySecurityGroupRuleUpdate,
		Delete: resourceScalewaySecurityGroupRuleDelete,
		Schema: map[string]*schema.Schema{
			"security_group": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The security group this rule is attached to",
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value != "accept" && value != "drop" {
						errors = append(errors, fmt.Errorf("%q must be one of 'accept', 'drop'", k))
					}
					return
				},
				Description: "The action to take when the security group rule is triggered (accept or drop)",
			},
			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value != "inbound" && value != "outbound" {
						errors = append(errors, fmt.Errorf("%q must be one of 'inbound', 'outbound'", k))
					}
					return
				},
				Description: "The direction the traffic is affected (inbound or outbound)",
			},
			"ip_range": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ip range affected by the security group rule",
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value != "ICMP" && value != "TCP" && value != "UDP" {
						errors = append(errors, fmt.Errorf("%q must be one of 'ICMP', 'TCP', 'UDP", k))
					}
					return
				},
				Description: "The protocol of the security group rule (ICMP, TCP or UDP)",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The port affected by the security group rule",
			},
		},
	}
}

func resourceScalewaySecurityGroupRuleCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	req := api.NewSecurityGroupRule{
		Action:       d.Get("action").(string),
		Direction:    d.Get("direction").(string),
		IPRange:      d.Get("ip_range").(string),
		Protocol:     d.Get("protocol").(string),
		DestPortFrom: d.Get("port").(int),
	}

	rule, err := scaleway.CreateSecurityGroupRule(d.Get("security_group").(string), req)
	if err != nil {
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error creating Security Group Rule: %q\n", serr.APIMessage)
		}

		return err
	}

	d.SetId(rule.ID)

	if d.Id() == "" {
		return fmt.Errorf("Failed to find created security group rule")
	}

	return resourceScalewaySecurityGroupRuleRead(d, m)
}

func resourceScalewaySecurityGroupRuleRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient
	rule, err := scaleway.GetSecurityGroupRule(d.Get("security_group").(string), d.Id())

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

	d.Set("action", rule.Action)
	d.Set("direction", rule.Direction)
	d.Set("ip_range", rule.IPRange)
	d.Set("protocol", rule.Protocol)
	d.Set("port", rule.DestPortFrom)

	return nil
}

func resourceScalewaySecurityGroupRuleUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	var req = api.UpdateSecurityGroupRule{
		Action:       d.Get("action").(string),
		Direction:    d.Get("direction").(string),
		IPRange:      d.Get("ip_range").(string),
		Protocol:     d.Get("protocol").(string),
		DestPortFrom: d.Get("port").(int),
	}

	if _, err := scaleway.UpdateSecurityGroupRule(req, d.Get("security_group").(string), d.Id()); err != nil {
		log.Printf("[DEBUG] error updating Security Group Rule: %q", err)

		return err
	}

	return resourceScalewaySecurityGroupRuleRead(d, m)
}

func resourceScalewaySecurityGroupRuleDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	err := scaleway.DeleteSecurityGroupRule(d.Get("security_group").(string), d.Id())
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
