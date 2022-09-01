package scaleway

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceScalewayInstanceSecurityGroupRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayInstanceSecurityGroupRulesCreate,
		ReadContext:   resourceScalewayInstanceSecurityGroupRulesRead,
		UpdateContext: resourceScalewayInstanceSecurityGroupRulesUpdate,
		DeleteContext: resourceScalewayInstanceSecurityGroupRulesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultInstanceSecurityGroupRuleTimeout),
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

// securityGroupRuleSchema returns schema for inbound/outbound rule in security group
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
				ValidateFunc: validation.IsIPAddress,
				Description:  "Ip address for this rule (e.g: 1.1.1.1). Only one of ip or ip_range should be provided",
			},
			"ip_range": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsCIDRNetwork(0, 128),
				Description:  "Ip range for this rule (e.g: 192.168.1.0/24). Only one of ip or ip_range should be provided",
			},
		},
	}
}

func resourceScalewayInstanceSecurityGroupRulesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("security_group_id").(string))

	// We call update instead of read as it will take care of creating rules.
	return resourceScalewayInstanceSecurityGroupRulesUpdate(ctx, d, meta)
}

func resourceScalewayInstanceSecurityGroupRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	securityGroupZonedID := d.Id()

	instanceAPI, zone, securityGroupID, err := instanceAPIWithZoneAndID(meta, securityGroupZonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("security_group_id", securityGroupZonedID)

	inboundRules, outboundRules, err := getSecurityGroupRules(ctx, instanceAPI, zone, securityGroupID, d)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("inbound_rule", inboundRules)
	_ = d.Set("outbound_rule", outboundRules)

	return nil
}

func resourceScalewayInstanceSecurityGroupRulesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	securityGroupZonedID := d.Id()
	instanceAPI, zone, securityGroupID, err := instanceAPIWithZoneAndID(meta, securityGroupZonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = updateSecurityGroupeRules(ctx, d, zone, securityGroupID, instanceAPI)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayInstanceSecurityGroupRulesRead(ctx, d, meta)
}

func resourceScalewayInstanceSecurityGroupRulesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	securityGroupZonedID := d.Id()
	instanceAPI, zone, securityGroupID, err := instanceAPIWithZoneAndID(meta, securityGroupZonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("inbound_rule", nil)
	_ = d.Set("outbound_rule", nil)

	err = updateSecurityGroupeRules(ctx, d, zone, securityGroupID, instanceAPI)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// updateSecurityGroupeRules handles updating SecurityGroupRules
//
// It works as followed:
//  1. Creates 2 map[direction][]rule: one for rules in state and one for rules in API nolint:gofmt
//  2. For each direction we:
//     A) Loop for each rule in state for this direction
//     a) Compare with api rule in this direction at the same index
//     if different update / if equals do nothing / if no more api rules to compare create new api rule
//     B) If there is more rule in the API we remove them
func updateSecurityGroupeRules(ctx context.Context, d *schema.ResourceData, zone scw.Zone, securityGroupID string, instanceAPI *instance.API) error {
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
	}, scw.WithAllPages(), scw.WithContext(ctx))
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
			stateRule, err := securityGroupRuleExpand(rawStateRule)
			if err != nil {
				return err
			}

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
				}, scw.WithContext(ctx))
				if err != nil {
					return err
				}
				continue
			}

			// We compare rule stateRule[index] and apiRule[index]. If they are different we update api rule to match state.
			apiRule := apiRules[direction][index]
			if ok, _ := securityGroupRuleEquals(stateRule, apiRule); !ok {
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
				}, scw.WithContext(ctx))
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
			}, scw.WithContext(ctx))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getSecurityGroupRules(ctx context.Context, instanceAPI *instance.API, zone scw.Zone, securityGroupID string, d *schema.ResourceData) ([]interface{}, []interface{}, error) {
	resRules, err := instanceAPI.ListSecurityGroupRules(&instance.ListSecurityGroupRulesRequest{
		Zone:            zone,
		SecurityGroupID: expandID(securityGroupID),
	}, scw.WithAllPages(), scw.WithContext(ctx))
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
				stateRule, errGroup := securityGroupRuleExpand(stateRules[direction][index])
				if errGroup != nil {
					return nil, nil, errGroup
				}
				if ok, _ := securityGroupRuleEquals(stateRule, apiRule); !ok {
					stateRules[direction][index], err = securityGroupRuleFlatten(apiRule)
					if err != nil {
						return nil, nil, err
					}
				}
			} else {
				rulesGroup, err := securityGroupRuleFlatten(apiRule)
				if err != nil {
					return nil, nil, err
				}
				stateRules[direction] = append(stateRules[direction], rulesGroup)
			}
		}
		// There are rule in tfstate not present in api
		if len(apiRules[direction]) != len(stateRules[direction]) {
			// Truncate stateRules with apiRules length
			stateRules[direction] = stateRules[direction][0:len(apiRules[direction])]
		}
	}

	return stateRules[instance.SecurityGroupRuleDirectionInbound], stateRules[instance.SecurityGroupRuleDirectionOutbound], nil
}

// securityGroupRuleExpand transform a state rule to an api one.
func securityGroupRuleExpand(i interface{}) (*instance.SecurityGroupRule, error) {
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

	ipnetRange, err := expandIPNet(ipRange)
	if err != nil {
		return nil, err
	}
	rule := &instance.SecurityGroupRule{
		DestPortFrom: &portFrom,
		DestPortTo:   &portTo,
		Protocol:     instance.SecurityGroupRuleProtocol(rawRule["protocol"].(string)),
		IPRange:      ipnetRange,
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

	return rule, nil
}

// securityGroupRuleFlatten transform an api rule to a state one.
func securityGroupRuleFlatten(rule *instance.SecurityGroupRule) (map[string]interface{}, error) {
	portFrom, portTo := uint32(0), uint32(0)

	if rule.DestPortFrom != nil {
		portFrom = *rule.DestPortFrom
	}

	if rule.DestPortTo != nil {
		portTo = *rule.DestPortTo
	}

	ipnetRange, err := flattenIPNet(rule.IPRange)
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{
		"protocol":   rule.Protocol.String(),
		"ip_range":   ipnetRange,
		"port_range": fmt.Sprintf("%d-%d", portFrom, portTo),
		"action":     rule.Action.String(),
	}
	return res, nil
}

// securityGroupRuleEquals compares two security group rule.
func securityGroupRuleEquals(ruleA, ruleB *instance.SecurityGroupRule) (bool, error) {
	zeroIfNil := func(v *uint32) uint32 {
		if v == nil {
			return 0
		}
		return *v
	}
	portFromEqual := zeroIfNil(ruleA.DestPortFrom) == zeroIfNil(ruleB.DestPortFrom)
	portToEqual := zeroIfNil(ruleA.DestPortTo) == zeroIfNil(ruleB.DestPortTo)
	ipRangeA, err := flattenIPNet(ruleA.IPRange)
	if err != nil {
		return false, err
	}
	ipRangeB, err := flattenIPNet(ruleB.IPRange)
	if err != nil {
		return false, err
	}
	ipEqual := ipRangeA == ipRangeB

	return ruleA.Action == ruleB.Action &&
		portFromEqual &&
		portToEqual &&
		ipEqual &&
		ruleA.Protocol == ruleB.Protocol, nil
}
