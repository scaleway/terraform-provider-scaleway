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

var resourceScalewayComputeInstanceSecurityGroupProtocols = []string{
	instance.SecurityGroupRuleProtocolICMP.String(),
	instance.SecurityGroupRuleProtocolTCP.String(),
	instance.SecurityGroupRuleProtocolUDP.String(),
}

var resourceScalewayComputeInstanceSecurityGroupRuleDirections = []string{
	instance.SecurityGroupRuleDirectionInbound.String(),
	instance.SecurityGroupRuleDirectionOutbound.String(),
}

var resourceScalewayComputeInstanceSecurityGroupActionReverse = map[instance.SecurityGroupPolicy]instance.SecurityGroupRuleAction{
	instance.SecurityGroupPolicyAccept: instance.SecurityGroupRuleActionDrop,
	instance.SecurityGroupPolicyDrop:   instance.SecurityGroupRuleActionAccept,
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
			"rule": {
				Type:        schema.TypeSet,
				Set:         securityGroupRuleHash,
				Optional:    true,
				Description: "inbound rules for this security group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      instance.SecurityGroupRuleDirectionInbound.String(),
							ValidateFunc: validation.StringInSlice(resourceScalewayComputeInstanceSecurityGroupRuleDirections, false),
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      instance.SecurityGroupRuleProtocolTCP.String(),
							ValidateFunc: validation.StringInSlice(resourceScalewayComputeInstanceSecurityGroupProtocols, false),
						},
						"port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"port_range": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip_range": {
							Type:     schema.TypeString,
							Optional: true,
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
	instanceApi, zone, err := getInstanceAPIWithZone(d, meta)
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
	instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(m, d.Id())
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

	//
	// Handle SecurityGroupRules
	//

	resRules, err := instanceApi.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
		Zone:            zone,
		SecurityGroupID: ID,
	}, scw.WithAllPages())
	if err != nil {
		return err
	}

	stateRules := d.Get("rule").(*schema.Set)
	apiRules := schema.NewSet(securityGroupRuleHash, nil)
	for _, rule := range resRules.Rules {
		if rule.Editable == false {
			continue
		}

		if rule.Action != securityGroupExpectedAction(rule, d) {
			continue
		}
		flat := securityGroupRuleFlatten(rule)
		apiRules.Add(flat)
	}

	rules := apiRules.Union(stateRules)
	d.Set("rule", rules)
	return nil
}

func resourceScalewayComputeInstanceSecurityGroupUpdate(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	inboundDefaultPolicy := instance.SecurityGroupPolicy(d.Get("inbound_default_policy").(string))
	outboundDefaultPolicy := instance.SecurityGroupPolicy(d.Get("outbound_default_policy").(string))

	updateReq := &instance.UpdateSecurityGroupRequest{
		Zone:                  zone,
		Name:                  utils.String(d.Get("name").(string)),
		SecurityGroupID:       ID,
		Description:           utils.String(d.Get("description").(string)),
		InboundDefaultPolicy:  &inboundDefaultPolicy,
		OutboundDefaultPolicy: &outboundDefaultPolicy,
	}

	_, err = instanceApi.UpdateSecurityGroup(updateReq)
	if err != nil {
		return err
	}

	//
	// Handle SecurityGroupRules
	//

	resRules, err := instanceApi.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
		Zone:            zone,
		SecurityGroupID: ID,
	}, scw.WithAllPages())
	if err != nil {
		return err
	}

	// Create two set of rule one with the target state and the other from api
	targetRules := schema.NewSet(securityGroupRuleWithActionHash, nil)
	apiRules := schema.NewSet(securityGroupRuleWithActionHash, nil)

	for _, rawRule := range d.Get("rule").(*schema.Set).List() {
		rule := securityGroupRuleExpand(rawRule)
		rule.Action = securityGroupExpectedAction(rule, d)
		targetRules.Add(securityGroupRuleFlatten(rule))
	}

	for _, rule := range resRules.Rules {
		if rule.Editable == false {
			continue
		}
		apiRules.Add(securityGroupRuleFlatten(rule))
	}

	// Using set we can get the rules to add and the rules to delete
	rulesToAdd := targetRules.Difference(apiRules)
	rulesToDel := apiRules.Difference(targetRules)

	for _, rawRule := range rulesToAdd.List() {
		rule := securityGroupRuleExpand(rawRule)

		_, err = instanceApi.CreateSecurityGroupRule(&instance.CreateSecurityGroupRuleRequest{
			Zone:            zone,
			SecurityGroupID: ID,
			Protocol:        rule.Protocol,
			DestPortFrom:    rule.DestPortFrom,
			DestPortTo:      rule.DestPortTo,
			Direction:       rule.Direction,
			Action:          rule.Action,
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

func portRangeFlatten(from, to uint32) string {
	if to == 0 {
		to = from
	}
	return fmt.Sprintf("%d-%d", from, to)
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

	portRange := rawRule["port_range"].(string)
	if portRange == "" {
		portRange = rawRule["port"].(string)
	}
	from, to := portRangeExpand(portRange)

	id, _ := rawRule["id"].(string)
	action, _ := rawRule["action"].(string)

	ipRange := rawRule["ip_range"].(string)
	if ipRange == "" {
		ipRange = rawRule["ip"].(string) + "/32"
	}
	if ipRange == "/32" {
		ipRange = "0.0.0.0/0"
	}

	return &instance.SecurityGroupRule{
		ID:           id,
		DestPortTo:   to,
		DestPortFrom: from,
		Protocol:     instance.SecurityGroupRuleProtocol(rawRule["protocol"].(string)),
		IPRange:      ipRange,
		Direction:    instance.SecurityGroupRuleDirection(rawRule["type"].(string)),
		Action:       instance.SecurityGroupRuleAction(action),
	}
}

func securityGroupRuleFlatten(rule *instance.SecurityGroupRule) map[string]interface{} {
	res := map[string]interface{}{
		"id":         rule.ID,
		"protocol":   rule.Protocol.String(),
		"ip_range":   ipv4RangeFormat(rule.IPRange),
		"port_range": portRangeFlatten(rule.DestPortFrom, rule.DestPortTo),
		"type":       rule.Direction.String(),
		"action":     rule.Action.String(),
	}
	return res
}

func securityGroupRuleHash(i interface{}) int {
	rule := securityGroupRuleExpand(i)
	s := fmt.Sprintf("%s/%s/%d-%d/%s", rule.Protocol.String(), rule.IPRange, rule.DestPortFrom, rule.DestPortFrom, rule.Direction)
	return schema.HashString(s)
}

func securityGroupRuleWithActionHash(i interface{}) int {
	rule := securityGroupRuleExpand(i)
	s := fmt.Sprintf("%d/%s", securityGroupRuleHash(i), rule.Action)
	return schema.HashString(s)
}

func securityGroupExpectedAction(rule *instance.SecurityGroupRule, d *schema.ResourceData) instance.SecurityGroupRuleAction {
	inboundDefaultPolicy := instance.SecurityGroupPolicy(d.Get("inbound_default_policy").(string))
	outboundDefaultPolicy := instance.SecurityGroupPolicy(d.Get("outbound_default_policy").(string))

	switch rule.Direction {
	case instance.SecurityGroupRuleDirectionInbound:
		return resourceScalewayComputeInstanceSecurityGroupActionReverse[inboundDefaultPolicy]
	default:
		return resourceScalewayComputeInstanceSecurityGroupActionReverse[outboundDefaultPolicy]
	}
}
