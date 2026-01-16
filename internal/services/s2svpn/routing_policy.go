package s2svpn

import (
	"context"
	_ "time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceRoutingPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceRoutingPolicyCreate,
		ReadContext:   ResourceRoutingPolicyRead,
		UpdateContext: ResourceRoutingPolicyUpdate,
		DeleteContext: ResourceRoutingPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
			Description: "The list of tags to apply to the routing policy",
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
			ForceNew:    true,
			Description: "IP prefixes to accept from the peer (ranges of route announcements to accept)",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
			},
		},
		"prefix_filter_out": {
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    true,
			Description: "IP prefix filters to advertise to the peer (ranges of routes to advertise)",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
			},
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the TLS stage",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the TLS stage",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Organization ID of the Project",
		},
	}
}

func ResourceRoutingPolicyCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
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

	req := &s2s_vpn.CreateRoutingPolicyRequest{
		Region:          region,
		ProjectID:       d.Get("project_id").(string),
		Name:            types.ExpandOrGenerateString(d.Get("name").(string), "connection"),
		Tags:            types.ExpandStrings(d.Get("tags")),
		IsIPv6:          d.Get("is_ipv6").(bool),
		PrefixFilterIn:  prefixFilterIn,
		PrefixFilterOut: prefixFilterOut,
	}

	res, err := api.CreateRoutingPolicy(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))

	return ResourceRoutingPolicyRead(ctx, d, m)
}

func ResourceRoutingPolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := api.GetRoutingPolicy(&s2s_vpn.GetRoutingPolicyRequest{
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

	prefixFilterIn, err := FlattenPrefixFilters(policy.PrefixFilterIn)
	if err != nil {
		return diag.FromErr(err)
	}

	prefixFilterOut, err := FlattenPrefixFilters(policy.PrefixFilterOut)
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

	req := &s2s_vpn.UpdateRoutingPolicyRequest{
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

	err = api.DeleteRoutingPolicy(&s2s_vpn.DeleteRoutingPolicyRequest{
		Region:          region,
		RoutingPolicyID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
