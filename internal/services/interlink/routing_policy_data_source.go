package interlink

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	interlink "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/routing_policy_data_source.md
var routingPolicyDataSourceDescription string

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
		ReadContext: DataSourceRoutingPolicyRead,
		Description: routingPolicyDataSourceDescription,
	}
}

func DataSourceRoutingPolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := interlinkAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	routingPolicyID, ok := d.GetOk("routing_policy_id")
	if !ok {
		policyName := d.Get("name").(string)

		res, err := api.ListRoutingPolicies(&interlink.ListRoutingPoliciesRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(policyName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundPolicy, err := datasource.FindExact(
			res.RoutingPolicies,
			func(s *interlink.RoutingPolicy) bool { return s.Name == policyName },
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

	policy, err := api.GetRoutingPolicy(&interlink.GetRoutingPolicyRequest{
		RoutingPolicyID: locality.ExpandID(routingPolicyID),
		Region:          region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setRoutingPolicyState(d, policy)
}
