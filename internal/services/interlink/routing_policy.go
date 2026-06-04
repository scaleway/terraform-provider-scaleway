package interlink

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/routing_policy.md
var routingPolicyDescription string

func ResourceRoutingPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceRoutingPolicyCreate,
		ReadContext:   ResourceRoutingPolicyRead,
		UpdateContext: ResourceRoutingPolicyUpdate,
		DeleteContext: ResourceRoutingPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description:   routingPolicyDescription,
		Identity:      identity.DefaultRegional(),
		SchemaVersion: 0,
		SchemaFunc:    routingPolicySchema,
	}
}

func routingPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The name of the routing policy",
		},
		"tags": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of tags associated with the routing policy",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"is_ipv6": {
			Type:        schema.TypeBool,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
			Description: "IP prefixes version of the routing policy",
		},
		"prefix_filter_in": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "IP prefixes to accept from the peer (ranges of route announcements to accept)",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
			},
		},
		"prefix_filter_out": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "IP prefix filters to advertise to the peer (ranges of routes to advertise)",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
			},
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the routing policy",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the routing policy",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The Organization ID the routing policy is associated with",
		},
	}
}

func ResourceRoutingPolicyCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	prefixFilterIn, err := expandPrefixFilters(d.Get("prefix_filter_in"))
	if err != nil {
		return diag.FromErr(err)
	}

	prefixFilterOut, err := expandPrefixFilters(d.Get("prefix_filter_out"))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &interlink.CreateRoutingPolicyRequest{
		Region:          region,
		ProjectID:       d.Get("project_id").(string),
		Name:            types.ExpandOrGenerateString(d.Get("name").(string), "routing-policy"),
		Tags:            types.ExpandStrings(d.Get("tags")),
		IsIPv6:          d.Get("is_ipv6").(bool),
		PrefixFilterIn:  prefixFilterIn,
		PrefixFilterOut: prefixFilterOut,
	}

	res, err := api.CreateRoutingPolicy(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetRegionalIdentity(d, res.Region, res.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceRoutingPolicyRead(ctx, d, m)
}

func ResourceRoutingPolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := api.GetRoutingPolicy(&interlink.GetRoutingPolicyRequest{
		RoutingPolicyID: id,
		Region:          region,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	diags := setRoutingPolicyState(d, policy)

	err = identity.SetRegionalIdentity(d, policy.Region, policy.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setRoutingPolicyState(d *schema.ResourceData, policy *interlink.RoutingPolicy) diag.Diagnostics {
	prefixFilterIn, err := flattenPrefixFilters(policy.PrefixFilterIn)
	if err != nil {
		return diag.FromErr(err)
	}

	prefixFilterOut, err := flattenPrefixFilters(policy.PrefixFilterOut)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", policy.Name)
	_ = d.Set("region", policy.Region)
	_ = d.Set("project_id", policy.ProjectID)
	_ = d.Set("organization_id", policy.OrganizationID)
	_ = d.Set("tags", policy.Tags)
	_ = d.Set("created_at", types.FlattenTime(policy.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(policy.UpdatedAt))
	_ = d.Set("is_ipv6", policy.IsIPv6)
	_ = d.Set("prefix_filter_in", prefixFilterIn)
	_ = d.Set("prefix_filter_out", prefixFilterOut)

	return nil
}

func ResourceRoutingPolicyUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	req := &interlink.UpdateRoutingPolicyRequest{
		Region:          region,
		RoutingPolicyID: id,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if d.HasChange("prefix_filter_in") {
		req.PrefixFilterIn = types.ExpandStringsPtr(d.Get("prefix_filter_in"))
		hasChanged = true
	}

	if d.HasChange("prefix_filter_out") {
		req.PrefixFilterOut = types.ExpandStringsPtr(d.Get("prefix_filter_out"))
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateRoutingPolicy(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceRoutingPolicyRead(ctx, d, m)
}

func ResourceRoutingPolicyDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteRoutingPolicy(&interlink.DeleteRoutingPolicyRequest{
		Region:          region,
		RoutingPolicyID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
