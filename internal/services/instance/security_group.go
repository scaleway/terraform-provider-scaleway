package instance

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceSecurityGroupCreate,
		ReadContext:   ResourceInstanceSecurityGroupRead,
		UpdateContext: ResourceInstanceSecurityGroupUpdate,
		DeleteContext: ResourceInstanceSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultInstanceSecurityGroupTimeout),
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "accept",
				Description:      "Default inbound traffic policy for this security group",
				ValidateDiagFunc: verify.ValidateEnum[instanceSDK.SecurityGroupPolicy](),
			},
			"outbound_default_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "accept",
				Description:      "Default outbound traffic policy for this security group",
				ValidateDiagFunc: verify.ValidateEnum[instanceSDK.SecurityGroupPolicy](),
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
			"enable_default_security": {
				Type:        schema.TypeBool,
				Description: "Enable blocking of SMTP on IPv4 and IPv6",
				Optional:    true,
				Default:     true,
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the security group",
			},
			"zone":            zonal.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
	}
}

func ResourceInstanceSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instanceSDK.CreateSecurityGroupRequest{
		Name:                  types.ExpandOrGenerateString(d.Get("name"), "sg"),
		Zone:                  zone,
		Project:               types.ExpandStringPtr(d.Get("project_id")),
		Description:           d.Get("description").(string),
		Stateful:              d.Get("stateful").(bool),
		InboundDefaultPolicy:  instanceSDK.SecurityGroupPolicy(d.Get("inbound_default_policy").(string)),
		OutboundDefaultPolicy: instanceSDK.SecurityGroupPolicy(d.Get("outbound_default_policy").(string)),
		EnableDefaultSecurity: types.ExpandBoolPtr(d.Get("enable_default_security")),
	}
	tags := types.ExpandStrings(d.Get("tags"))

	if len(tags) > 0 {
		req.Tags = tags
	}

	res, err := instanceAPI.CreateSecurityGroup(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, res.SecurityGroup.ID))

	if d.Get("external_rules").(bool) {
		return ResourceInstanceSecurityGroupRead(ctx, d, m)
	}
	// We call update instead of read as it will take care of creating rules.
	return ResourceInstanceSecurityGroupUpdate(ctx, d, m)
}

func ResourceInstanceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.GetSecurityGroup(&instanceSDK.GetSecurityGroupRequest{
		SecurityGroupID: ID,
		Zone:            zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("zone", zone)
	_ = d.Set("organization_id", res.SecurityGroup.Organization)
	_ = d.Set("project_id", res.SecurityGroup.Project)
	_ = d.Set("name", res.SecurityGroup.Name)
	_ = d.Set("stateful", res.SecurityGroup.Stateful)
	_ = d.Set("description", res.SecurityGroup.Description)
	_ = d.Set("inbound_default_policy", res.SecurityGroup.InboundDefaultPolicy.String())
	_ = d.Set("outbound_default_policy", res.SecurityGroup.OutboundDefaultPolicy.String())
	_ = d.Set("enable_default_security", res.SecurityGroup.EnableDefaultSecurity)
	_ = d.Set("tags", res.SecurityGroup.Tags)

	if !d.Get("external_rules").(bool) {
		inboundRules, outboundRules, err := getSecurityGroupRules(ctx, instanceAPI, zone, ID, d)
		if err != nil {
			return diag.FromErr(err)
		}

		_ = d.Set("inbound_rule", inboundRules)
		_ = d.Set("outbound_rule", outboundRules)
	}

	return nil
}

func getSecurityGroupRules(ctx context.Context, instanceAPI *instanceSDK.API, zone scw.Zone, securityGroupID string, d *schema.ResourceData) ([]interface{}, []interface{}, error) {
	resRules, err := instanceAPI.ListSecurityGroupRules(&instanceSDK.ListSecurityGroupRulesRequest{
		Zone:            zone,
		SecurityGroupID: locality.ExpandID(securityGroupID),
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return nil, nil, err
	}

	sort.Slice(resRules.Rules, func(i, j int) bool {
		return resRules.Rules[i].Position < resRules.Rules[j].Position
	})

	apiRules := map[instanceSDK.SecurityGroupRuleDirection][]*instanceSDK.SecurityGroupRule{
		instanceSDK.SecurityGroupRuleDirectionInbound:  {},
		instanceSDK.SecurityGroupRuleDirectionOutbound: {},
	}

	stateRules := map[instanceSDK.SecurityGroupRuleDirection][]interface{}{
		instanceSDK.SecurityGroupRuleDirectionInbound:  d.Get("inbound_rule").([]interface{}),
		instanceSDK.SecurityGroupRuleDirectionOutbound: d.Get("outbound_rule").([]interface{}),
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

				if ok, _ := SecurityGroupRuleEquals(stateRule, apiRule); !ok {
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

	return stateRules[instanceSDK.SecurityGroupRuleDirectionInbound], stateRules[instanceSDK.SecurityGroupRuleDirectionOutbound], nil
}

func ResourceInstanceSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, _, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, ID, err := zonal.ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	inboundDefaultPolicy := instanceSDK.SecurityGroupPolicy("")
	if d.Get("inbound_default_policy") != nil {
		inboundDefaultPolicy = instanceSDK.SecurityGroupPolicy(d.Get("inbound_default_policy").(string))
	}

	outboundDefaultPolicy := instanceSDK.SecurityGroupPolicy("")
	if d.Get("outbound_default_policy") != nil {
		outboundDefaultPolicy = instanceSDK.SecurityGroupPolicy(d.Get("outbound_default_policy").(string))
	}

	description := ""
	if d.Get("description") != nil {
		description = d.Get("description").(string)
	}

	updateReq := &instanceSDK.UpdateSecurityGroupRequest{
		Zone:                  zone,
		SecurityGroupID:       ID,
		Stateful:              scw.BoolPtr(d.Get("stateful").(bool)),
		Description:           types.ExpandStringPtr(description),
		InboundDefaultPolicy:  inboundDefaultPolicy,
		OutboundDefaultPolicy: outboundDefaultPolicy,
		Tags:                  scw.StringsPtr([]string{}),
	}

	tags := types.ExpandStrings(d.Get("tags"))
	if len(tags) > 0 {
		updateReq.Tags = scw.StringsPtr(types.ExpandStrings(d.Get("tags")))
	}

	if d.HasChange("enable_default_security") {
		updateReq.EnableDefaultSecurity = types.ExpandBoolPtr(d.Get("enable_default_security"))
	}

	// Only update name if one is provided in the state
	if d.Get("name") != nil && d.Get("name").(string) != "" {
		updateReq.Name = types.ExpandStringPtr(d.Get("name"))
	}

	_, err = instanceAPI.UpdateSecurityGroup(updateReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if !d.Get("external_rules").(bool) {
		err = updateSecurityGroupeRules(ctx, d, zone, ID, instanceAPI)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceInstanceSecurityGroupRead(ctx, d, m)
}

// updateSecurityGroupeRules handles updating SecurityGroupRules
func updateSecurityGroupeRules(ctx context.Context, d *schema.ResourceData, zone scw.Zone, securityGroupID string, instanceAPI *instanceSDK.API) error {
	stateRules := map[instanceSDK.SecurityGroupRuleDirection][]interface{}{
		instanceSDK.SecurityGroupRuleDirectionInbound:  d.Get("inbound_rule").([]interface{}),
		instanceSDK.SecurityGroupRuleDirectionOutbound: d.Get("outbound_rule").([]interface{}),
	}

	setGroupRules := []*instanceSDK.SetSecurityGroupRulesRequestRule{}

	for direction := range stateRules {
		// Loop for all state rules in this direction
		for _, rawStateRule := range stateRules[direction] {
			stateRule, err := securityGroupRuleExpand(rawStateRule)
			if err != nil {
				return err
			}

			// This happens when there is more rule in state than in the api. We create more rule in API.
			setGroupRules = append(setGroupRules, &instanceSDK.SetSecurityGroupRulesRequestRule{
				Zone:         &zone,
				Protocol:     stateRule.Protocol,
				IPRange:      stateRule.IPRange,
				Action:       stateRule.Action,
				DestPortTo:   stateRule.DestPortTo,
				DestPortFrom: stateRule.DestPortFrom,
				Direction:    direction,
			})
		}
	}

	_, err := instanceAPI.SetSecurityGroupRules(&instanceSDK.SetSecurityGroupRulesRequest{
		SecurityGroupID: securityGroupID,
		Zone:            zone,
		Rules:           setGroupRules,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	return nil
}

func ResourceInstanceSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, _, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, ID, err := zonal.ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = instanceAPI.DeleteSecurityGroup(&instanceSDK.DeleteSecurityGroupRequest{
		SecurityGroupID: ID,
		Zone:            zone,
	}, scw.WithContext(ctx))

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

// securityGroupRuleSchema returns schema for inbound/outbound rule in security group
func securityGroupRuleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: verify.ValidateEnum[instanceSDK.SecurityGroupRuleAction](),
				Description:      "Action when rule match request (drop or accept)",
			},
			"protocol": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          instanceSDK.SecurityGroupRuleProtocolTCP.String(),
				ValidateDiagFunc: verify.ValidateEnum[instanceSDK.SecurityGroupRuleProtocol](),
				Description:      "Protocol for this rule (TCP, UDP, ICMP or ANY)",
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
				Deprecated:   "Ip address is deprecated. Please use ip_range instead",
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

// securityGroupRuleExpand transform a state rule to an api one.
func securityGroupRuleExpand(i interface{}) (*instanceSDK.SecurityGroupRule, error) {
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

	ipnetRange, err := types.ExpandIPNet(ipRange)
	if err != nil {
		return nil, err
	}

	rule := &instanceSDK.SecurityGroupRule{
		DestPortFrom: &portFrom,
		DestPortTo:   &portTo,
		Protocol:     instanceSDK.SecurityGroupRuleProtocol(rawRule["protocol"].(string)),
		IPRange:      ipnetRange,
		Action:       instanceSDK.SecurityGroupRuleAction(action),
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
func securityGroupRuleFlatten(rule *instanceSDK.SecurityGroupRule) (map[string]interface{}, error) {
	portFrom, portTo := uint32(0), uint32(0)

	if rule.DestPortFrom != nil {
		portFrom = *rule.DestPortFrom
	}

	if rule.DestPortTo != nil {
		portTo = *rule.DestPortTo
	}

	ipnetRange, err := types.FlattenIPNet(rule.IPRange)
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

// SecurityGroupRuleEquals compares two security group rule.
func SecurityGroupRuleEquals(ruleA, ruleB *instanceSDK.SecurityGroupRule) (bool, error) {
	zeroIfNil := func(v *uint32) uint32 {
		if v == nil {
			return 0
		}

		return *v
	}
	portFromEqual := zeroIfNil(ruleA.DestPortFrom) == zeroIfNil(ruleB.DestPortFrom)
	portToEqual := zeroIfNil(ruleA.DestPortTo) == zeroIfNil(ruleB.DestPortTo)

	ipRangeA, err := types.FlattenIPNet(ruleA.IPRange)
	if err != nil {
		return false, err
	}

	ipRangeB, err := types.FlattenIPNet(ruleB.IPRange)
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
