package scaleway

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayInstanceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayInstanceSecurityGroupCreate,
		Read:   resourceScalewayInstanceSecurityGroupRead,
		Update: resourceScalewayInstanceSecurityGroupUpdate,
		Delete: resourceScalewayInstanceSecurityGroupDelete,
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
			"stateful": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "The stateful value of the security group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the security group",
			},
			"inbound_default_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "accept",
				Description: "Default inbound traffic policy for this security group",
				ValidateFunc: validation.StringInSlice([]string{
					instance.SecurityGroupPolicyAccept.String(),
					instance.SecurityGroupPolicyDrop.String(),
				}, false),
			},
			"outbound_default_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "accept",
				Description: "Default outbound traffic policy for this security group",
				ValidateFunc: validation.StringInSlice([]string{
					instance.SecurityGroupPolicyAccept.String(),
					instance.SecurityGroupPolicyDrop.String(),
				}, false),
			},
			"inbound_rule": {
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "Inbound rules for this security group",
				Elem:          securityGroupRuleSchema(),
				ConflictsWith: []string{"external_rules"},
			},
			"outbound_rule": {
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "Outbound rules for this security group",
				Elem:          securityGroupRuleSchema(),
				ConflictsWith: []string{"external_rules"},
			},
			"external_rules": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"inbound_rule", "outbound_rule"},
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayInstanceSecurityGroupCreate(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return err
	}

	res, err := instanceApi.CreateSecurityGroup(&instance.CreateSecurityGroupRequest{
		Name:                  expandOrGenerateString(d.Get("name"), "sg"),
		Zone:                  zone,
		Organization:          expandStringPtr(d.Get("organization_id")),
		Description:           d.Get("description").(string),
		Stateful:              d.Get("stateful").(bool),
		InboundDefaultPolicy:  instance.SecurityGroupPolicy(d.Get("inbound_default_policy").(string)),
		OutboundDefaultPolicy: instance.SecurityGroupPolicy(d.Get("outbound_default_policy").(string)),
	})
	if err != nil {
		return err
	}

	d.SetId(newZonedId(zone, res.SecurityGroup.ID))

	if d.Get("external_rules").(bool) {
		return resourceScalewayInstanceSecurityGroupRead(d, m)
	} else {
		// We call update instead of read as it will take care of creating rules.
		return resourceScalewayInstanceSecurityGroupUpdate(d, m)
	}
}

func resourceScalewayInstanceSecurityGroupRead(d *schema.ResourceData, m interface{}) error {
	instanceApi, zone, ID, err := instanceAPIWithZoneAndID(m, d.Id())
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

	_ = d.Set("zone", zone)
	_ = d.Set("organization_id", res.SecurityGroup.Organization)
	_ = d.Set("name", res.SecurityGroup.Name)
	_ = d.Set("stateful", res.SecurityGroup.Stateful)
	_ = d.Set("description", res.SecurityGroup.Description)
	_ = d.Set("inbound_default_policy", res.SecurityGroup.InboundDefaultPolicy.String())
	_ = d.Set("outbound_default_policy", res.SecurityGroup.OutboundDefaultPolicy.String())

	if !d.Get("external_rules").(bool) {
		inboundRules, outboundRules, err := getSecurityGroupRules(instanceApi, zone, ID, d)
		if err != nil {
			return err
		}
		_ = d.Set("inbound_rule", inboundRules)
		_ = d.Set("outbound_rule", outboundRules)
	}
	return nil
}

func getSecurityGroupRules(instanceApi *instance.API, zone scw.Zone, securityGroupID string, d *schema.ResourceData) ([]interface{}, []interface{}, error) {

	resRules, err := instanceApi.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
		Zone:            zone,
		SecurityGroupID: expandID(securityGroupID),
	}, scw.WithAllPages())
	if err != nil {
		return nil, nil, err
	}
	sort.Slice(resRules.Rules, func(i, j int) bool {
		return resRules.Rules[i].Position < resRules.Rules[j].Position
	})
	apiRules := map[instance.SecurityGroupRuleDirection][]*instance.SecurityGroupRule{
		instance.SecurityGroupRuleDirectionInbound:  {},
		instance.SecurityGroupRuleDirectionOutbound: {},
	}

	stateRules := map[instance.SecurityGroupRuleDirection][]interface{}{
		instance.SecurityGroupRuleDirectionInbound:  d.Get("inbound_rule").([]interface{}),
		instance.SecurityGroupRuleDirectionOutbound: d.Get("outbound_rule").([]interface{}),
	}

	for _, apiRule := range resRules.Rules {
		if !apiRule.Editable {
			continue
		}
		apiRules[apiRule.Direction] = append(apiRules[apiRule.Direction], apiRule)
	}

	// We make sure that we keep state rule if they match their api rule.
	for direction := range apiRules {
		for index, apiRule := range apiRules[direction] {
			if index < len(stateRules[direction]) {
				stateRule := securityGroupRuleExpand(stateRules[direction][index])
				if !securityGroupRuleEquals(stateRule, apiRule) {
					stateRules[direction][index] = securityGroupRuleFlatten(apiRule)
				}
			} else {
				stateRules[direction] = append(stateRules[direction], securityGroupRuleFlatten(apiRule))
			}
		}
	}

	return stateRules[instance.SecurityGroupRuleDirectionInbound], stateRules[instance.SecurityGroupRuleDirectionOutbound], nil
}

func resourceScalewayInstanceSecurityGroupUpdate(d *schema.ResourceData, m interface{}) error {
	meta := m.(*Meta)
	instanceApi := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(d.Id())
	if err != nil {
		return err
	}

	inboundDefaultPolicy := instance.SecurityGroupPolicy("")
	if d.Get("inbound_default_policy") != nil {
		inboundDefaultPolicy = instance.SecurityGroupPolicy(d.Get("inbound_default_policy").(string))
	}
	outboundDefaultPolicy := instance.SecurityGroupPolicy("")
	if d.Get("outbound_default_policy") != nil {
		outboundDefaultPolicy = instance.SecurityGroupPolicy(d.Get("outbound_default_policy").(string))
	}

	description := ""
	if d.Get("description") != nil {
		description = d.Get("description").(string)
	}
	updateReq := &instance.UpdateSecurityGroupRequest{
		Zone:                  zone,
		SecurityGroupID:       ID,
		Stateful:              scw.BoolPtr(d.Get("stateful").(bool)),
		Description:           scw.StringPtr(description),
		InboundDefaultPolicy:  &inboundDefaultPolicy,
		OutboundDefaultPolicy: &outboundDefaultPolicy,
	}

	// Only update name if one is provided in the state
	if d.Get("name") != nil && d.Get("name").(string) != "" {
		updateReq.Name = scw.StringPtr(d.Get("name").(string))
	}

	_, err = instanceApi.UpdateSecurityGroup(updateReq)
	if err != nil {
		return err
	}

	if !d.Get("external_rules").(bool) {
		err = updateSecurityGroupeRules(d, zone, ID, instanceApi)
		if err != nil {
			return err
		}
	}

	return resourceScalewayInstanceSecurityGroupRead(d, m)
}

// updateSecurityGroupeRules handles updating SecurityGroupRules
//
// It works as followed:
//   1) Creates 2 map[direction][]rule: one for rules in state and one for rules in API
//   2) For each direction we:
//     A) Loop for each rule in state for this direction
//       a) Compare with api rule in this direction at the same index
//          if different update / if equals do nothing / if no more api rules to compare create new api rule
//     B) If there is more rule in the API we remove them
func updateSecurityGroupeRules(d *schema.ResourceData, zone scw.Zone, securityGroupID string, instanceAPI *instance.API) error {
	apiRules := map[instance.SecurityGroupRuleDirection][]*instance.SecurityGroupRule{
		instance.SecurityGroupRuleDirectionInbound:  {},
		instance.SecurityGroupRuleDirectionOutbound: {},
	}
	stateRules := map[instance.SecurityGroupRuleDirection][]interface{}{
		instance.SecurityGroupRuleDirectionInbound:  d.Get("inbound_rule").([]interface{}),
		instance.SecurityGroupRuleDirectionOutbound: d.Get("outbound_rule").([]interface{}),
	}

	// Fill apiRules with data from API
	resRules, err := instanceAPI.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
		Zone:            zone,
		SecurityGroupID: expandID(securityGroupID),
	}, scw.WithAllPages())
	if err != nil {
		return err
	}
	sort.Slice(resRules.Rules, func(i, j int) bool {
		return resRules.Rules[i].Position < resRules.Rules[j].Position
	})
	for _, apiRule := range resRules.Rules {
		if !apiRule.Editable {
			continue
		}
		apiRules[apiRule.Direction] = append(apiRules[apiRule.Direction], apiRule)
	}

	// Loop through all directions
	for direction := range stateRules {

		// Loop for all state rules in this direction
		for index, rawStateRule := range stateRules[direction] {
			apiRule := (*instance.SecurityGroupRule)(nil)
			stateRule := securityGroupRuleExpand(rawStateRule)

			// This happen when there is more rule in state than in the api. We create more rule in API.
			if index >= len(apiRules[direction]) {
				_, err = instanceAPI.CreateSecurityGroupRule(&instance.CreateSecurityGroupRuleRequest{
					Zone:            zone,
					SecurityGroupID: securityGroupID,
					Protocol:        stateRule.Protocol,
					IPRange:         stateRule.IPRange,
					Action:          stateRule.Action,
					DestPortTo:      stateRule.DestPortTo,
					DestPortFrom:    stateRule.DestPortFrom,
					Direction:       direction,
				})
				if err != nil {
					return err
				}
				continue
			}

			// We compare rule stateRule[index] and apiRule[index]. If they are different we update api rule to match state.
			apiRule = apiRules[direction][index]
			if !securityGroupRuleEquals(stateRule, apiRule) {
				destPortFrom := stateRule.DestPortFrom
				destPortTo := stateRule.DestPortTo
				if destPortFrom == nil {
					destPortFrom = scw.Uint32Ptr(0)
				}
				if destPortTo == nil {
					destPortTo = scw.Uint32Ptr(0)
				}

				_, err = instanceAPI.UpdateSecurityGroupRule(&instance.UpdateSecurityGroupRuleRequest{
					Zone:                zone,
					SecurityGroupID:     securityGroupID,
					SecurityGroupRuleID: apiRule.ID,
					Protocol:            &stateRule.Protocol,
					IPRange:             &stateRule.IPRange,
					Action:              &stateRule.Action,
					DestPortTo:          destPortTo,
					DestPortFrom:        destPortFrom,
					Direction:           &direction,
				})
				if err != nil {
					return err
				}
			}
		}

		// We loop through remaining API rules and delete them as they are no longer in the state.
		for index := len(stateRules[direction]); index < len(apiRules[direction]); index++ {
			err = instanceAPI.DeleteSecurityGroupRule(&instance.DeleteSecurityGroupRuleRequest{
				Zone:                zone,
				SecurityGroupID:     securityGroupID,
				SecurityGroupRuleID: apiRules[direction][index].ID,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func resourceScalewayInstanceSecurityGroupDelete(d *schema.ResourceData, m interface{}) error {
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

// securityGroupRuleSchema returns schema for inboud/outbout rule in security group
func securityGroupRuleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					instance.SecurityGroupRuleActionAccept.String(),
					instance.SecurityGroupRuleActionDrop.String(),
				}, false),
				Description: "Action when rule match request (drop or accept)",
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  instance.SecurityGroupRuleProtocolTCP.String(),
				ValidateFunc: validation.StringInSlice([]string{
					instance.SecurityGroupRuleProtocolICMP.String(),
					instance.SecurityGroupRuleProtocolTCP.String(),
					instance.SecurityGroupRuleProtocolUDP.String(),
					instance.SecurityGroupRuleProtocolANY.String(),
				}, false),
				Description: "Protocol for this rule (TCP, UDP, ICMP or ANY)",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Network port for this rule",
			},
			"port_range": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Computed port range for this rule (e.g: 1-1024, 22-22)",
			},
			"ip": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.SingleIP(),
				Description:  "Ip address for this rule (e.g: 1.1.1.1). Only one of ip or ip_range should be provided",
			},
			"ip_range": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.CIDRNetwork(0, 128),
				Description:  "Ip range for this rule (e.g: 192.168.1.0/24). Only one of ip or ip_range should be provided",
			},
		},
	}
}

// securityGroupRuleExpand transform a state rule to an api one.
func securityGroupRuleExpand(i interface{}) *instance.SecurityGroupRule {
	rawRule := i.(map[string]interface{})

	portFrom, portTo := uint32(0), uint32(0)

	portRange := rawRule["port_range"].(string)
	if portRange != "" {
		_, _ = fmt.Sscanf(portRange, "%d-%d", &portFrom, &portTo)
	} else {
		portFrom = uint32(rawRule["port"].(int))
		portTo = portFrom
	}

	action, _ := rawRule["action"].(string)
	ipRange := rawRule["ip_range"].(string)
	if ipRange == "" {
		ipRange = rawRule["ip"].(string) + "/32"
	}
	if ipRange == "/32" {
		ipRange = "0.0.0.0/0"
	}

	rule := &instance.SecurityGroupRule{
		DestPortFrom: &portFrom,
		DestPortTo:   &portTo,
		Protocol:     instance.SecurityGroupRuleProtocol(rawRule["protocol"].(string)),
		IPRange:      expandIPNet(ipRange),
		Action:       instance.SecurityGroupRuleAction(action),
	}

	if *rule.DestPortFrom == *rule.DestPortTo {
		rule.DestPortTo = nil
	}

	// Handle when no port is specified.
	if portFrom == 0 && portTo == 0 {
		rule.DestPortFrom = nil
		rule.DestPortTo = nil
	}

	return rule
}

// securityGroupRuleFlatten transform a api rule to an state one.
func securityGroupRuleFlatten(rule *instance.SecurityGroupRule) map[string]interface{} {
	portFrom, portTo := uint32(0), uint32(0)

	if rule.DestPortFrom != nil {
		portFrom = *rule.DestPortFrom
	}

	if rule.DestPortTo != nil {
		portTo = *rule.DestPortTo
	}

	res := map[string]interface{}{
		"protocol":   rule.Protocol.String(),
		"ip_range":   flattenIpNet(rule.IPRange),
		"port_range": fmt.Sprintf("%d-%d", portFrom, portTo),
		"action":     rule.Action.String(),
	}
	return res
}

// securityGroupRuleEquals compares two security group rule.
func securityGroupRuleEquals(ruleA, ruleB *instance.SecurityGroupRule) bool {
	zeroIfNil := func(v *uint32) uint32 {
		if v == nil {
			return 0
		}
		return *v
	}
	portFromEqual := zeroIfNil(ruleA.DestPortFrom) == zeroIfNil(ruleB.DestPortFrom)
	portToEqual := zeroIfNil(ruleA.DestPortTo) == zeroIfNil(ruleB.DestPortTo)
	ipEqual := flattenIpNet(ruleA.IPRange) == flattenIpNet(ruleB.IPRange)

	return ruleA.Action == ruleB.Action &&
		portFromEqual &&
		portToEqual &&
		ipEqual &&
		ruleA.Protocol == ruleB.Protocol
}
