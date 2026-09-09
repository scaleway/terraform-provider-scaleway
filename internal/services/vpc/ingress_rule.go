package vpc

import (
	"context"
	_ "embed"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/ingress_rule_resource.md
var ingressRuleResourceDescription string

func ResourceIngressRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceIngressRuleCreate,
		ReadContext:   ResourceIngressRuleRead,
		UpdateContext: ResourceIngressRuleUpdate,
		DeleteContext: ResourceIngressRuleDelete,
		Description:   ingressRuleResourceDescription,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    ingressRuleSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func ingressRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"vpc_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			Description:      "The ID of the VPC the ingress rule belongs to",
			DiffSuppressFunc: dsf.Locality,
		},
		"source": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Source IP range to which this rule applies (CIDR notation with subnet mask)",
		},
		"nexthop_resource_ip": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "IP of the nexthop resource for the ingress rule",
		},
		"nexthop_private_network_id": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "The ID of the nexthop private network",
			DiffSuppressFunc: dsf.Locality,
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The ingress rule description",
		},
		"tags": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The tags associated with the ingress rule",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"region": regional.Schema(),
		// Computed elements
		"is_ipv6": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether the ingress rule is for IPv6 traffic",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the ingress rule",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the ingress rule",
		},
		"srn": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The Scaleway Resource Name (SRN) of the ingress rule",
		},
	}
}

func ResourceIngressRuleCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	source, err := types.ExpandIPNet(d.Get("source").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpc.CreateIngressRuleRequest{
		Region:                  region,
		VpcID:                   locality.ExpandID(d.Get("vpc_id").(string)),
		Source:                  source,
		NexthopPrivateNetworkID: locality.ExpandID(d.Get("nexthop_private_network_id").(string)),
		NexthopResourceIP:       net.ParseIP(d.Get("nexthop_resource_ip").(string)),
		Description:             types.ExpandStringPtr(d.Get("description")),
		Tags:                    types.ExpandStrings(d.Get("tags")),
	}

	res, err := vpcAPI.CreateIngressRule(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))

	err = identity.SetRegionalIdentity(d, region, res.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceIngressRuleRead(ctx, d, m)
}

func ResourceIngressRuleRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcAPI.GetIngressRule(&vpc.GetIngressRuleRequest{
		Region: region,
		RuleID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	diags := setIngressRuleState(d, res, region)

	err = identity.SetRegionalIdentity(d, region, ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setIngressRuleState(d *schema.ResourceData, rule *vpc.IngressRule, region scw.Region) diag.Diagnostics {
	source, err := types.FlattenIPNet(rule.Source)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("vpc_id", regional.NewIDString(region, rule.VpcID))
	_ = d.Set("source", source)
	_ = d.Set("is_ipv6", rule.IsIPv6)

	if rule.NexthopResourceIP != nil {
		_ = d.Set("nexthop_resource_ip", rule.NexthopResourceIP.String())
	} else {
		_ = d.Set("nexthop_resource_ip", "")
	}

	_ = d.Set("nexthop_private_network_id", regional.NewIDString(region, rule.NexthopPrivateNetworkID))
	_ = d.Set("description", types.FlattenStringPtr(rule.Description))
	_ = d.Set("tags", rule.Tags)
	_ = d.Set("region", region)
	_ = d.Set("created_at", types.FlattenTime(rule.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(rule.UpdatedAt))
	_ = d.Set("srn", rule.Srn)

	return nil
}

func ResourceIngressRuleUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	updateRequest := &vpc.UpdateIngressRuleRequest{
		Region: region,
		RuleID: ID,
	}

	if d.HasChange("source") {
		source, err := types.ExpandIPNet(d.Get("source").(string))
		if err != nil {
			return diag.FromErr(err)
		}

		updateRequest.Source = &source
		hasChanged = true
	}

	if d.HasChange("nexthop_resource_ip") {
		ipStr := d.Get("nexthop_resource_ip").(string)
		updateRequest.NexthopResourceIP = new(net.ParseIP(ipStr))
		hasChanged = true
	}

	if d.HasChange("nexthop_private_network_id") {
		updateRequest.NexthopPrivateNetworkID = types.ExpandUpdatedStringPtr(locality.ExpandID(d.Get("nexthop_private_network_id")))
		hasChanged = true
	}

	if d.HasChange("description") {
		updateRequest.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if hasChanged {
		_, err = vpcAPI.UpdateIngressRule(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceIngressRuleRead(ctx, d, m)
}

func ResourceIngressRuleDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = vpcAPI.DeleteIngressRule(&vpc.DeleteIngressRuleRequest{
		Region: region,
		RuleID: ID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
