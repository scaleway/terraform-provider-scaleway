package s2svpn

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	s2svpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceRoutingPolicy() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceRoutingPolicy().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"routing_policy_id"}
	dsSchema["routing_policy_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the routing policy",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceS2SRoutingPolicyRead,
	}
}

func DataSourceS2SRoutingPolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	routingPolicyID, ok := d.GetOk("routing_policy_id")
	if !ok {
		policyName := d.Get("name").(string)

		res, err := api.ListRoutingPolicies(&s2svpn.ListRoutingPoliciesRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(policyName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundPolicy, err := datasource.FindExact(
			res.RoutingPolicies,
			func(s *s2svpn.RoutingPolicy) bool { return s.Name == policyName },
			policyName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		routingPolicyID = foundPolicy.ID
	}

	regionalID := datasource.NewRegionalID(routingPolicyID, region)
	d.SetId(regionalID)
	_ = d.Set("routing_policy_id", regionalID)

	diags := ResourceRoutingPolicyRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read routing policy state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("routing policy (%s) not found", regionalID)
	}

	return nil
}
