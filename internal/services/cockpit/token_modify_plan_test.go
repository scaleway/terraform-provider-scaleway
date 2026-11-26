package cockpit_test

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/cockpit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenModifyPlan_OmittedScopesAfterImport(t *testing.T) {
	ctx := t.Context()
	r := &cockpit.TokenResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), schemaResp.Diagnostics)

	customScopes := tokenScopesList(t, map[string]bool{
		"query_metrics": true,
		"write_metrics": true,
	})

	state := newTokenState(t, ctx, schemaResp.Schema, customScopes, types.StringNull())
	plan := newTokenPlan(t, ctx, schemaResp.Schema, types.ListNull(tokenScopesAttrTypes()), types.StringNull())

	resp := &resource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{State: state, Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	assert.Empty(t, resp.RequiresReplace)
}

func TestTokenModifyPlan_ExplicitScopesChangeRequiresReplace(t *testing.T) {
	ctx := t.Context()
	r := &cockpit.TokenResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), schemaResp.Diagnostics)

	oldScopes := tokenScopesList(t, map[string]bool{"write_metrics": true, "write_logs": false})
	newScopes := tokenScopesList(t, map[string]bool{"write_metrics": false, "write_logs": true})

	state := newTokenState(t, ctx, schemaResp.Schema, oldScopes, types.StringValue("secret"))
	plan := newTokenPlan(t, ctx, schemaResp.Schema, newScopes, types.StringValue("secret"))

	resp := &resource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{State: state, Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.Len(t, resp.RequiresReplace, 1)
	assert.True(t, resp.RequiresReplace[0].Equal(path.Root("scopes")))
}

func tokenScopesAttrTypes() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"query_metrics":       types.BoolType,
		"write_metrics":       types.BoolType,
		"setup_metrics_rules": types.BoolType,
		"query_logs":          types.BoolType,
		"write_logs":          types.BoolType,
		"setup_logs_rules":    types.BoolType,
		"setup_alerts":        types.BoolType,
		"query_traces":        types.BoolType,
		"write_traces":        types.BoolType,
	}}
}

func tokenScopesList(t *testing.T, enabled map[string]bool) types.List {
	t.Helper()

	defaults := map[string]bool{
		"query_metrics":       false,
		"write_metrics":       true,
		"setup_metrics_rules": false,
		"query_logs":          false,
		"write_logs":          true,
		"setup_logs_rules":    false,
		"setup_alerts":        false,
		"query_traces":        false,
		"write_traces":        false,
	}
	maps.Copy(defaults, enabled)

	values := make(map[string]attr.Value, len(defaults))
	for name, value := range defaults {
		values[name] = types.BoolValue(value)
	}

	object, diags := types.ObjectValue(tokenScopesAttrTypes().AttrTypes, values)
	require.False(t, diags.HasError(), diags)

	list, diags := types.ListValue(tokenScopesAttrTypes(), []attr.Value{object})
	require.False(t, diags.HasError(), diags)

	return list
}

type tokenAttrTarget interface {
	SetAttribute(context.Context, path.Path, any) diag.Diagnostics
}

func newTokenState(t *testing.T, ctx context.Context, sch schema.Schema, scopes types.List, secretKey types.String) tfsdk.State {
	t.Helper()

	state := tfsdk.State{
		Schema: sch,
		Raw:    tftypes.NewValue(sch.Type().TerraformType(ctx), nil),
	}
	setTokenAttrs(t, ctx, &state, scopes, secretKey)

	return state
}

func newTokenPlan(t *testing.T, ctx context.Context, sch schema.Schema, scopes types.List, secretKey types.String) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{
		Schema: sch,
		Raw:    tftypes.NewValue(sch.Type().TerraformType(ctx), nil),
	}
	setTokenAttrs(t, ctx, &plan, scopes, secretKey)

	return plan
}

func setTokenAttrs(t *testing.T, ctx context.Context, target tokenAttrTarget, scopes types.List, secretKey types.String) {
	t.Helper()

	timeoutsNull := timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"delete": types.StringType,
	})}

	attrs := map[string]any{
		"timeouts":   timeoutsNull,
		"scopes":     scopes,
		"id":         types.StringValue("fr-par/3ed23100-398f-4b39-b3ab-fadd1b693da5"),
		"name":       types.StringValue("tok"),
		"secret_key": secretKey,
		"created_at": types.StringNull(),
		"updated_at": types.StringNull(),
		"project_id": types.StringValue("11111111-1111-1111-1111-111111111111"),
		"region":     types.StringValue("fr-par"),
	}

	for name, value := range attrs {
		diags := target.SetAttribute(ctx, path.Root(name), value)
		require.False(t, diags.HasError(), diags)
	}
}
