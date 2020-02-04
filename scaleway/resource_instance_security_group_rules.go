package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
)

func resourceScalewayInstanceSecurityGroupRules() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayInstanceSecurityGroupRulesCreate,
		Read:   resourceScalewayInstanceSecurityGroupRulesRead,
		Update: resourceScalewayInstanceSecurityGroupRulesUpdate,
		Delete: resourceScalewayInstanceSecurityGroupRulesDelete,
		Importer: &schema.ResourceImporter{
			State: customImporterState,
		},

		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				// Ensure SecurityGroupRules.ID and SecurityGroupRules.security_group_id stay in sync.
				// If security_group_id is changed, a new SecurityGroupRules is created, with a new ID.
				ForceNew:    true,
				Description: "The security group associated with this volume",
			},
			"inbound_rule": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Inbound rules for this set of security group rules",
				Elem:        securityGroupRuleSchema(),
			},
			"outbound_rule": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Outbound rules for this set of security group rules",
				Elem:        securityGroupRuleSchema(),
			},
		},
	}
}

func customImporterState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// importZID is the ID set by the user when using the Terraform's import function.
	// It should be a SecurityGroupZID.
	importZID := d.Id()

	d.Set("security_group_id", importZID)
	d.SetId(securityGroupRulesZIDFromSecurityGroupZID(importZID))

	return []*schema.ResourceData{d}, nil
}

func resourceScalewayInstanceSecurityGroupRulesCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId(securityGroupRulesZIDFromSecurityGroupZID(d.Get("security_group_id").(string)))

	// We call update instead of read as it will take care of creating rules.
	return resourceScalewayInstanceSecurityGroupRulesUpdate(d, m)
}

func resourceScalewayInstanceSecurityGroupRulesRead(d *schema.ResourceData, m interface{}) error {
	securityGroupRulesZID := d.Id()
	securityGroupZID := securityGroupZIDFromSecurityGroupRulesZID(securityGroupRulesZID)

	instanceApi, zone, securityGroupID, err := instanceAPIWithZoneAndID(m, securityGroupZID)
	if err != nil {
		return err
	}

	d.Set("security_group_id", securityGroupZID)

	inboundRules, outboundRules, err := getSecurityGroupRules(instanceApi, zone, securityGroupID, d)
	if err != nil {
		return err
	}

	d.Set("inbound_rule", inboundRules)
	d.Set("outbound_rule", outboundRules)

	return nil
}

func resourceScalewayInstanceSecurityGroupRulesUpdate(d *schema.ResourceData, m interface{}) error {
	securityGroupZID := securityGroupZIDFromSecurityGroupRulesZID(d.Id())
	instanceApi, zone, securityGroupID, err := instanceAPIWithZoneAndID(m, securityGroupZID)
	if err != nil {
		return err
	}

	err = updateSecurityGroupeRules(d, zone, securityGroupID, instanceApi)
	if err != nil {
		return err
	}

	return resourceScalewayInstanceSecurityGroupRulesRead(d, m)
}

func resourceScalewayInstanceSecurityGroupRulesDelete(d *schema.ResourceData, m interface{}) error {
	securityGroupZID := securityGroupZIDFromSecurityGroupRulesZID(d.Id())
	instanceApi, zone, securityGroupID, err := instanceAPIWithZoneAndID(m, securityGroupZID)
	if err != nil {
		return err
	}

	d.Set("inbound_rule", nil)
	d.Set("outbound_rule", nil)

	err = updateSecurityGroupeRules(d, zone, securityGroupID, instanceApi)
	if err != nil {
		return err
	}

	return nil
}

// securityGroupRulesZIDFromSecurityGroupZID returns the
//
// SecurityGroupRules.ID should be the based on SecurityGroup.ID.
// This is necessary to support Terraform import feature. -> false -> why is this necessary ?
// If we want to support multiple SGs(SecurityGroup) for 1 SGRS(SecurityGroupRules),
// we could always use the first SecurityGroup's ID,
// because from the API,
// the data for a single SGRS is duplicated for all SGs using the same SGRS.
func securityGroupRulesZIDFromSecurityGroupZID(SecurityGroupZID string) string {
	return securityGroupZIDFromSecurityGroupRulesZID(SecurityGroupZID) + "-sgrs-id"
}

func securityGroupZIDFromSecurityGroupRulesZID(zid string) string {
	return strings.Replace(zid, "-sgrs-id", "", 1)
}
