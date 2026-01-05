package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceVPCACLCreate,
		ReadContext:   ResourceVPCACLRead,
		UpdateContext: ResourceVPCACLUpdate,
		DeleteContext: ResourceVPCACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    aclSchema,
		Identity:      identity.DefaultRegional(),
	}
}

func aclSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"vpc_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The VPC in which to create the ACL rule",
		},
		"default_policy": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          vpc.ActionAccept,
			Description:      "The action to take for packets which do not match any rules",
			ValidateDiagFunc: verify.ValidateEnum[vpc.Action](),
		},
		"is_ipv6": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Defines whether this set of ACL rules is for IPv6 (false = IPv4). Each Network ACL can have rules for only one IP type",
		},
		"rules": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of Network ACL rules",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"protocol": {
						Type:             schema.TypeString,
						Optional:         true,
						Default:          "ANY",
						Description:      "The protocol to which this rule applies. Default value: ANY",
						ValidateDiagFunc: verify.ValidateEnum[vpc.ACLRuleProtocol](),
					},
					"source": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Source IP range to which this rule applies (CIDR notation with subnet mask)",
					},
					"src_port_low": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Starting port of the source port range to which this rule applies (inclusive)",
					},
					"src_port_high": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Ending port of the source port range to which this rule applies (inclusive)",
					},
					"destination": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Destination IP range to which this rule applies (CIDR notation with subnet mask)",
					},
					"dst_port_low": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Starting port of the destination port range to which this rule applies (inclusive)",
					},
					"dst_port_high": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Ending port of the destination port range to which this rule applies (inclusive)",
					},
					"action": {
						Type:             schema.TypeString,
						Optional:         true,
						Description:      "The policy to apply to the packet",
						ValidateDiagFunc: verify.ValidateEnum[vpc.Action](),
					},
					"description": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The rule description",
					},
				},
			},
		},
		"region": regional.Schema(),
	}
}

func ResourceVPCACLCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpc.SetACLRequest{
		VpcID:         locality.ExpandID(d.Get("vpc_id").(string)),
		IsIPv6:        d.Get("is_ipv6").(bool),
		DefaultPolicy: vpc.Action(d.Get("default_policy").(string)),
		Region:        region,
	}

	expandedRules, err := expandACLRules(d.Get("rules"))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("rules") != nil {
		req.Rules = expandedRules
	}

	_, err = vpcAPI.SetACL(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetRegionalIdentity(d, region, regional.ExpandID(d.Get("vpc_id").(string)).ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCACLRead(ctx, d, m)
}

func ResourceVPCACLRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	acl, err := vpcAPI.GetACL(&vpc.GetACLRequest{
		VpcID:  locality.ExpandID(ID),
		Region: region,
		IsIPv6: d.Get("is_ipv6").(bool),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("rules", flattenACLRules(acl.Rules))
	_ = d.Set("default_policy", acl.DefaultPolicy.String())

	err = identity.SetRegionalIdentity(d, region, regional.ExpandID(d.Get("vpc_id").(string)).ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceVPCACLUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpc.SetACLRequest{
		VpcID:         locality.ExpandID(ID),
		IsIPv6:        d.Get("is_ipv6").(bool),
		DefaultPolicy: vpc.Action(d.Get("default_policy").(string)),
		Region:        region,
	}

	expandedRules, err := expandACLRules(d.Get("rules"))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("rules") != nil {
		req.Rules = expandedRules
	}

	_, err = vpcAPI.SetACL(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCACLRead(ctx, d, m)
}

func ResourceVPCACLDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = vpcAPI.SetACL(&vpc.SetACLRequest{
		VpcID:         locality.ExpandID(ID),
		Region:        region,
		DefaultPolicy: vpc.ActionAccept,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
