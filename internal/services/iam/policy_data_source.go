package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourcePolicy() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePolicy().Schema)
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

	diags := resourceIamPolicyRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read iam policy state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("iam policy (%s) not found", policyID)
	}

	return nil
}
