package scaleway

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewaySecurityGroupRule() *schema.Resource {
	return &schema.Resource{
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
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	req := api.NewSecurityGroupRule{
		Action:       d.Get("action").(string),
		Direction:    d.Get("direction").(string),
		IPRange:      d.Get("ip_range").(string),
		Protocol:     d.Get("protocol").(string),
		DestPortFrom: d.Get("port").(int),
	}

	var (
		rule *api.SecurityGroupRule
		err  error
	)
	if err = retry(func() error {
		rule, err = scaleway.PostSecurityGroupRule(d.Get("security_group").(string), req)
		if serr, ok := err.(api.APIError); ok {
			log.Printf("[DEBUG] Error creating Security Group Rule: %q\n", serr.APIMessage)
		}

		return err
	}); err != nil {
		return err
	}

	d.SetId(rule.ID)

	if d.Id() == "" {
		return fmt.Errorf("Failed to find created security group rule")
	}

	return resourceScalewaySecurityGroupRuleRead(d, m)
}

func resourceScalewaySecurityGroupRuleRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway
	var (
		rule *api.GetSecurityGroupRule
		err  error
	)
	if err = retry(func() error {
		rule, err = scaleway.GetASecurityGroupRule(d.Get("security_group").(string), d.Id())
		return err
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

	d.Set("action", rule.Rules.Action)
	d.Set("direction", rule.Rules.Direction)
	d.Set("ip_range", rule.Rules.IPRange)
	d.Set("protocol", rule.Rules.Protocol)
	d.Set("port", rule.Rules.DestPortFrom)

	return nil
}

func resourceScalewaySecurityGroupRuleUpdate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	var req = api.NewSecurityGroupRule{
		Action:       d.Get("action").(string),
		Direction:    d.Get("direction").(string),
		IPRange:      d.Get("ip_range").(string),
		Protocol:     d.Get("protocol").(string),
		DestPortFrom: d.Get("port").(int),
	}

	if err := retry(func() error {
		return scaleway.PutSecurityGroupRule(req, d.Get("security_group").(string), d.Id())
	}); err != nil {
		log.Printf("[DEBUG] error updating Security Group Rule: %q", err)

		return err
	}

	return resourceScalewaySecurityGroupRuleRead(d, m)
}

func resourceScalewaySecurityGroupRuleDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Client).scaleway

	mu.Lock()
	defer mu.Unlock()

	if err := retry(func() error {
		return scaleway.DeleteSecurityGroupRule(d.Get("security_group").(string), d.Id())
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
