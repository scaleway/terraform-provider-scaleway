package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"strings"
)

func resourceScalewayInstanceSecurityGroupRules() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayInstanceSecurityGroupRulesCreate,
		Read:   resourceScalewayInstanceSecurityGroupRulesRead,
		Update: resourceScalewayInstanceSecurityGroupRulesUpdate,
		Delete: resourceScalewayInstanceSecurityGroupRulesDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:        schema.TypeString,
				Required:    true,
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
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayInstanceSecurityGroupRulesCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId(securityGroupRulesIDFromSecurityGroupID(d))

	// We call update instead of read as it will take care of creating rules.
	return resourceScalewayInstanceSecurityGroupUpdate(d, m)
}

func resourceScalewayInstanceSecurityGroupRulesRead(d *schema.ResourceData, m interface{}) error {

	securityGroupRulesZoneID := d.Id()
	securityGroupZoneID := securityGroupZIDFromsecurityGroupRulesZID(securityGroupRulesZoneID)

	instanceApi, zone, securityGroupID, err := instanceAPIWithZoneAndID(m, securityGroupZoneID)
	if err != nil {
		return err
	}

	res, err := instanceApi.GetSecurityGroup(&instance.GetSecurityGroupRequest{
		SecurityGroupID: securityGroupID,
		Zone:            zone,
	})
	if err != nil {
		return err
	}

	d.Set("security_group_id", res.SecurityGroup.ID)
	d.Set("zone", zone)
	d.Set("organization_id", res.SecurityGroup.Organization)

	stateRules, err := getSecurityGroupRules(instanceApi, zone, securityGroupID, d)
	if err != nil {
		return err
	}

	d.Set("inbound_rule", stateRules[instance.SecurityGroupRuleDirectionInbound])
	d.Set("outbound_rule", stateRules[instance.SecurityGroupRuleDirectionOutbound])
	return nil
}

func resourceScalewayInstanceSecurityGroupRulesUpdate(d *schema.ResourceData, m interface{}) error {

	instanceApi, zone, securityGroupID, err := instanceAPIWithZoneAndID(m, securityGroupZoneIDFromData(d))
	if err != nil {
		return err
	}
	d.SetId(newZonedId(zone, securityGroupRulesIDFromSecurityGroupID(d)))

	err = updateSecurityGroupeRules(d, zone, securityGroupID, instanceApi)
	if err != nil {
		return err
	}

	return resourceScalewayInstanceSecurityGroupRulesRead(d, m)
}

func resourceScalewayInstanceSecurityGroupRulesDelete(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, securityGroupID, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	err = updateSecurityGroupeRules(d, zone, securityGroupID, instanceApi)
	if err != nil {
		return err
	}

	return nil
}

// securityGroupRulesIDZFromSecurityGroupIDZ returns the
//
// SecurityGroupRules.ID should be the based on SecurityGroup.ID.
// This is necessary to support Terraform import feature.
// If we want to support multiple SGs(SecurityGroup) for 1 SGRS(SecurityGroupRules),
// we could always use the first SecurityGroup's ID,
// because from the API,
// the data for a single SGRS is duplicated for all SGs using the same SGRS.
func securityGroupRulesIDFromSecurityGroupID(d *schema.ResourceData) string {
	// TODO: have different IDs for SecurityGroup and SecurityGroupRules
	// Adding the suffix generates an error because the ID is not a valid UUID
	// Can we disable that check ?
	// We should not have the same ID for SecurityGroup and SecurityGroupRules.
	return d.Get("security_group_id").(string) // + "-sgrs-id"
}

func securityGroupZIDFromsecurityGroupRulesZID(zid string) string {
	return strings.Replace(zid, "-sgrs-id", "", 1)
}

func securityGroupZoneIDFromData(d *schema.ResourceData) string {
	return securityGroupZIDFromsecurityGroupRulesZID(d.Id())
}
