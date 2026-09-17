package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourcePolicy() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePolicy().SchemaFunc())
	datasource.AddOptionalFieldsToSchema(dsSchema, "name")

	dsSchema["name"].ConflictsWith = []string{"policy_id"}
	dsSchema["policy_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the policy",
		ValidateDiagFunc: verify.IsUUID(),
	}

	return &schema.Resource{
		ReadContext: DataSourceIamPolicyRead,
		Schema:      dsSchema,
	}
}

func DataSourceIamPolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	iamAPI := NewAPI(m)

	policyID, policyIDExists := d.GetOk("policy_id")
	if !policyIDExists {
		policyName := d.Get("name").(string)

		res, err := iamAPI.ListPolicies(&iam.ListPoliciesRequest{
			PolicyName: types.ExpandStringPtr(policyName),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundPolicy, err := datasource.FindExact(
			res.Policies,
			func(s *iam.Policy) bool { return s.Name == policyName },
			policyName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		policyID = foundPolicy.ID
	}

	d.SetId(policyID.(string))

	err := d.Set("policy_id", policyID)
	if err != nil {
		return diag.FromErr(err)
	}

	pol, err := iamAPI.GetPolicy(&iam.GetPolicyRequest{
		PolicyID: policyID.(string),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	listRules, err := iamAPI.ListRules(&iam.ListRulesRequest{
		PolicyID: pol.ID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list policy's rules: %w", err))
	}

	setPolicyState(d, pol, listRules.Rules)

	return nil
}
