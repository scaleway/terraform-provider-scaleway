package vpc

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/ingress_rule_data_source.md
var ingressRuleDataSourceDescription string

func DataSourceIngressRule() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceIngressRule().SchemaFunc())

	filterFields := []string{"vpc_id", "nexthop_resource_ip", "nexthop_private_network_id", "is_ipv6", "tags"}

	dsSchema["ingress_rule_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the VPC ingress rule",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    filterFields,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "vpc_id", "nexthop_resource_ip", "nexthop_private_network_id", "tags", "region")

	for _, key := range []string{"vpc_id", "nexthop_resource_ip", "nexthop_private_network_id", "tags"} {
		dsSchema[key].ConflictsWith = []string{"ingress_rule_id"}
	}

	dsSchema["is_ipv6"] = &schema.Schema{
		Type:          schema.TypeBool,
		Optional:      true,
		Computed:      true,
		Description:   "Only ingress rules with the matching IP version will be returned",
		ConflictsWith: []string{"ingress_rule_id"},
	}

	return &schema.Resource{
		ReadContext: DataSourceIngressRuleRead,
		Description: ingressRuleDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceIngressRuleRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	ruleID, idExists := d.GetOk("ingress_rule_id")
	if idExists {
		return dataSourceIngressRuleReadByID(ctx, d, m, ruleID.(string))
	}

	return dataSourceIngressRuleReadByFilters(ctx, d, m)
}

func dataSourceIngressRuleReadByID(ctx context.Context, d *schema.ResourceData, m any, ruleID string) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	id := locality.ExpandID(ruleID)
	d.SetId(regional.NewIDString(region, id))

	rule, err := vpcAPI.GetIngressRule(&vpc.GetIngressRuleRequest{
		Region: region,
		RuleID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setIngressRuleState(d, rule, region)
}

func dataSourceIngressRuleReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpc.ListIngressRulesRequest{
		Region: region,
		Tags:   types.ExpandStrings(d.Get("tags")),
	}

	if vpcID, ok := d.GetOk("vpc_id"); ok {
		req.VpcID = types.ExpandStringPtr(locality.ExpandID(vpcID))
	}

	if pnID, ok := d.GetOk("nexthop_private_network_id"); ok {
		req.NexthopPrivateNetworkID = types.ExpandStringPtr(locality.ExpandID(pnID))
	}

	if rawIP, ok := d.GetOk("nexthop_resource_ip"); ok {
		req.NexthopResourceIP = new(net.ParseIP(rawIP.(string)))
	}

	if isIPv6, ok := d.GetOk("is_ipv6"); ok {
		req.IsIPv6 = types.ExpandBoolPtr(isIPv6)
	}

	res, err := vpcAPI.ListIngressRules(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Rules) == 0 {
		return diag.FromErr(errors.New("no VPC ingress rule found matching the specified filters"))
	}

	if len(res.Rules) > 1 {
		return diag.FromErr(fmt.Errorf("multiple VPC ingress rules (%d) found, please refine your filters or use ingress_rule_id", len(res.Rules)))
	}

	rule := res.Rules[0]
	d.SetId(regional.NewIDString(region, rule.ID))
	_ = d.Set("ingress_rule_id", regional.NewIDString(region, rule.ID))

	return setIngressRuleState(d, rule, region)
}
