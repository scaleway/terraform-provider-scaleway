//nolint:testpackage // Tests need access to unexported token model and scope helpers.
package cockpit

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenModifyPlan_OmittedScopesAfterImport(t *testing.T) {
	ctx := t.Context()
	r := &TokenResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), schemaResp.Diagnostics)

	customScopes, diags := flattenTokenScopes([]cockpit.TokenScope{
		cockpit.TokenScopeReadOnlyMetrics,
		cockpit.TokenScopeWriteOnlyMetrics,
	})
	require.False(t, diags.HasError(), diags)

	timeoutsNull := timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"delete": types.StringType,
	})}

	stateModel := tokenResourceModel{
		Timeouts:  timeoutsNull,
		Scopes:    customScopes,
		ID:        types.StringValue("fr-par/3ed23100-398f-4b39-b3ab-fadd1b693da5"),
		Name:      types.StringValue("imported"),
		SecretKey: types.StringNull(),
		ProjectID: types.StringValue("11111111-1111-1111-1111-111111111111"),
		Region:    types.StringValue("fr-par"),
	}
	planModel := stateModel
	planModel.Scopes = types.ListNull(tokenScopesObjectType())

	var state tfsdk.State

	state.Schema = schemaResp.Schema
	require.False(t, state.Set(ctx, &stateModel).HasError())

	var plan tfsdk.Plan

	plan.Schema = schemaResp.Schema
	require.False(t, plan.Set(ctx, &planModel).HasError())

	resp := &resource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{State: state, Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	assert.Empty(t, resp.RequiresReplace)
}

func TestTokenModifyPlan_ExplicitScopesChangeRequiresReplace(t *testing.T) {
	ctx := t.Context()
	r := &TokenResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), schemaResp.Diagnostics)

	oldScopes, diags := flattenTokenScopes([]cockpit.TokenScope{cockpit.TokenScopeWriteOnlyMetrics})
	require.False(t, diags.HasError(), diags)

	newScopes, diags := flattenTokenScopes([]cockpit.TokenScope{cockpit.TokenScopeWriteOnlyLogs})
	require.False(t, diags.HasError(), diags)

	timeoutsNull := timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"delete": types.StringType,
	})}

	stateModel := tokenResourceModel{
		Timeouts:  timeoutsNull,
		Scopes:    oldScopes,
		ID:        types.StringValue("fr-par/3ed23100-398f-4b39-b3ab-fadd1b693da5"),
		Name:      types.StringValue("tok"),
		SecretKey: types.StringValue("secret"),
		ProjectID: types.StringValue("11111111-1111-1111-1111-111111111111"),
		Region:    types.StringValue("fr-par"),
	}
	planModel := stateModel
	planModel.Scopes = newScopes

	var state tfsdk.State

	state.Schema = schemaResp.Schema
	require.False(t, state.Set(ctx, &stateModel).HasError())

	var plan tfsdk.Plan

	plan.Schema = schemaResp.Schema
	require.False(t, plan.Set(ctx, &planModel).HasError())

	resp := &resource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{State: state, Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.Len(t, resp.RequiresReplace, 1)
	assert.True(t, resp.RequiresReplace[0].Equal(path.Root("scopes")))
}
