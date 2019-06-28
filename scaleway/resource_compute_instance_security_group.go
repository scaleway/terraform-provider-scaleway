package scaleway

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

var resourceScalewayComputeInstanceSecurityGroupPolicies = []string{
	instance.SecurityGroupPolicyAccept.String(),
	instance.SecurityGroupPolicyDrop.String(),
}

var resourceScalewayComputeInstanceSecurityGroupProtocol = []string{
	instance.SecurityGroupRuleProtocolICMP.String(),
	instance.SecurityGroupRuleProtocolTCP.String(),
	instance.SecurityGroupRuleProtocolUDP.String(),
}

func resourceScalewayComputeInstanceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayComputeInstanceSecurityGroupCreate,
		Read:   resourceScalewayComputeInstanceSecurityGroupRead,
		Update: resourceScalewayComputeInstanceSecurityGroupUpdate,
		Delete: resourceScalewayComputeInstanceSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the security group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the security group",
			},
			"inbound_default_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "drop",
				Description:  "Default inbound traffic policy for this security group",
				ValidateFunc: validation.StringInSlice(resourceScalewayComputeInstanceSecurityGroupPolicies, false),
			},
			"outbound_default_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "accept",
				Description:  "Default outbound traffic policy for this security group",
				ValidateFunc: validation.StringInSlice(resourceScalewayComputeInstanceSecurityGroupPolicies, false),
			},
			"inbound_rule": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "inbound rules for this security group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      instance.SecurityGroupRuleProtocolTCP.String(),
							ValidateFunc: validation.StringInSlice(resourceScalewayComputeInstanceSecurityGroupProtocol, false),
						},
						"port_range": {
							Type:      schema.TypeString,
							Required:  true,
							StateFunc: portRangeFormat,
						},
						"ip_range": {
							Type:      schema.TypeString,
							Required:  true,
							StateFunc: ipv4RangeFormat,
						},
					},
				},
			},
			"zone":       zoneSchema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayComputeInstanceSecurityGroupCreate(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, err := getZone(d, meta)
	if err != nil {
		return err
	}

	projectID, err := getProjectId(d, meta)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	if name == "" {
		name = getRandomName("sg")
	}

	res, err := instanceApi.CreateSecurityGroup(&instance.CreateSecurityGroupRequest{
		Name:                  name,
		Zone:                  zone,
		Organization:          projectID,
		Description:           d.Get("description").(string),
		Stateful:              true,
		InboundDefaultPolicy:  instance.SecurityGroupPolicy(d.Get("inbound_default_policy").(string)),
		OutboundDefaultPolicy: instance.SecurityGroupPolicy(d.Get("outbound_default_policy").(string)),
	})
	if err != nil {
		return err
	}

	d.SetId(newZonedId(zone, res.SecurityGroup.ID))
	return resourceScalewayComputeInstanceSecurityGroupUpdate(d, m)
}

func resourceScalewayComputeInstanceSecurityGroupRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	res, err := instanceApi.GetSecurityGroup(&instance.GetSecurityGroupRequest{
		SecurityGroupID: ID,
		Zone:            zone,
	})
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("zone", zone)
	d.Set("project_id", res.SecurityGroup.Organization)
	d.Set("name", res.SecurityGroup.Name)
	d.Set("description", res.SecurityGroup.Description)
	d.Set("inbound_default_policy", res.SecurityGroup.InboundDefaultPolicy.String())
	d.Set("outbound_default_policy", res.SecurityGroup.OutboundDefaultPolicy.String())

	resRules, err := instanceApi.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
		Zone:            zone,
		SecurityGroupID: ID,
	}, scw.WithAllPages())
	if err != nil {
		return err
	}

	inboundRules := ([]interface{})(nil)
	for _, rule := range resRules.Rules {
		if rule.Direction == instance.SecurityGroupRuleDirectionOutbound {
			continue
		}
		inboundRules = append(inboundRules, securityGroupRuleFlatten(rule))
	}
	d.Set("inbound_rule", inboundRules)

	return nil
}

func portRangeFlatten(from, to uint32) string {
	if to == 0 {
		to = from
	}
	return fmt.Sprintf("%d-%d", from, to)
}

func portRangeFormat(i interface{}) string {
	return portRangeFlatten(portRangeExpand(i))
}

func ipv4RangeFormat(i interface{}) string {
	ipRange := i.(string)
	if !strings.Contains(ipRange, "/") {
		ipRange = ipRange + "/32"
	}
	return ipRange
}

func portRangeExpand(i interface{}) (uint32, uint32) {
	portRange := i.(string)

	var from, to uint32
	var err error

	switch {
	case portRange == "":
		return 0, 0
	case strings.Contains(portRange, "-"):
		_, err = fmt.Sscanf(portRange, "%d-%d", &from, &to)
	default:
		_, err = fmt.Sscanf(portRange, "%d", &from)
	}

	if err != nil {
		return 0, 0
	}

	if to == 0 {
		to = from
	}
	return from, to
}

func securityGroupRuleExpand(i interface{}) *instance.SecurityGroupRule {
	rawRule := i.(map[string]interface{})
	from, to := portRangeExpand(rawRule["port_range"])

	return &instance.SecurityGroupRule{
		ID:           rawRule["id"].(string),
		DestPortTo:   to,
		DestPortFrom: from,
		Protocol:     instance.SecurityGroupRuleProtocol(rawRule["protocol"].(string)),
		IPRange:      rawRule["ip_range"].(string),
	}
}

func securityGroupRuleFlatten(rule *instance.SecurityGroupRule) map[string]interface{} {
	res := map[string]interface{}{
		"id":         rule.ID,
		"protocol":   rule.Protocol.String(),
		"ip_range":   ipv4RangeFormat(rule.IPRange),
		"port_range": portRangeFlatten(rule.DestPortFrom, rule.DestPortTo),
	}
	return res
}

func securityGroupRuleHash(rule interface{}) int {
	r := rule.(map[string]interface{})
	s := fmt.Sprintf("%s/%s/%s", r["protocol"], r["ip_range"], r["port_range"])
	hash := schema.HashString(s)
	fmt.Println("HASSSINNNxNGGG => ", s, hash)
	return hash

}

func resourceScalewayComputeInstanceSecurityGroupUpdate(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	updateReq := &instance.UpdateSecurityGroupRequest{
		Zone:            zone,
		SecurityGroupID: ID,
	}

	if d.HasChange("name") {
		updateReq.Name = utils.String(d.Get("name").(string))
	}

	if d.HasChange("description") {
		updateReq.Description = utils.String(d.Get("description").(string))
	}

	if d.HasChange("inbound_default_policy") {
		inboundDefaultPolicy := instance.SecurityGroupPolicy(d.Get("inbound_default_policy").(string))
		updateReq.InboundDefaultPolicy = &inboundDefaultPolicy
	}

	if d.HasChange("outbound_default_policy") {
		outboundDefaultPolicy := instance.SecurityGroupPolicy(d.Get("outbound_default_policy").(string))
		updateReq.OutboundDefaultPolicy = &outboundDefaultPolicy
	}

	_, err = instanceApi.UpdateSecurityGroup(updateReq)
	if err != nil {
		return err
	}

	// Rules
	resRules, err := instanceApi.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
		Zone:            zone,
		SecurityGroupID: ID,
	}, scw.WithAllPages())
	if err != nil {
		return err
	}

	inboundRules := schema.NewSet(securityGroupRuleHash, d.Get("inbound_rule").([]interface{}))
	apiInboundRules := schema.NewSet(securityGroupRuleHash, nil)
	for _, rule := range resRules.Rules {
		if rule.Direction == instance.SecurityGroupRuleDirectionOutbound {
			continue
		}
		apiInboundRules.Add(securityGroupRuleFlatten(rule))
	}

	rulesToAdd := inboundRules.Difference(apiInboundRules)
	rulesToDel := apiInboundRules.Difference(inboundRules)

	for _, rawRule := range rulesToAdd.List() {
		rule := securityGroupRuleExpand(rawRule)

		_, err = instanceApi.CreateSecurityGroupRule(&instance.CreateSecurityGroupRuleRequest{
			Zone:            zone,
			SecurityGroupID: ID,
			Protocol:        rule.Protocol,
			DestPortFrom:    rule.DestPortFrom,
			DestPortTo:      rule.DestPortTo,
			Direction:       instance.SecurityGroupRuleDirectionInbound,
			Action:          instance.SecurityGroupRuleActionAccept,
			IPRange:         rule.IPRange,
		})
		if err != nil {
			return err
		}
	}

	for _, rawRule := range rulesToDel.List() {
		rule := securityGroupRuleExpand(rawRule)

		err = instanceApi.DeleteSecurityGroupRule(&instance.DeleteSecurityGroupRuleRequest{
			Zone:            zone,
			SecurityGroupID: ID,
			SecurityRuleID:  rule.ID,
		})
		if err != nil {
			return err
		}
	}

	return resourceScalewayComputeInstanceSecurityGroupRead(d, m)
}

func resourceScalewayComputeInstanceSecurityGroupDelete(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	err = instanceApi.DeleteSecurityGroup(&instance.DeleteSecurityGroupRequest{
		SecurityGroupID: ID,
		Zone:            zone,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
