package cockpit_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/cockpit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeTokenStateV0(t *testing.T) {
	ctx := t.Context()

	schemaResp := &resource.SchemaResponse{}
	(&cockpit.TokenResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)

	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(`{"id":"3ed23100-398f-4b39-b3ab-fadd1b693da5","name":"tok","secret_key":"secret"}`)},
	}
	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}

	upgraders := (&cockpit.TokenResource{}).UpgradeState(ctx)
	require.Contains(t, upgraders, int64(0))
	upgraders[0].StateUpgrader(ctx, req, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	var (
		id, region, secretKey types.String
		scopes                types.List
	)

	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.False(t, resp.State.GetAttribute(ctx, path.Root("region"), &region).HasError())
	require.False(t, resp.State.GetAttribute(ctx, path.Root("secret_key"), &secretKey).HasError())
	require.False(t, resp.State.GetAttribute(ctx, path.Root("scopes"), &scopes).HasError())

	assert.Equal(t, "fr-par/3ed23100-398f-4b39-b3ab-fadd1b693da5", id.ValueString())
	assert.Equal(t, "fr-par", region.ValueString())
	assert.Equal(t, "secret", secretKey.ValueString())
	assert.False(t, scopes.IsNull())
	assert.Empty(t, scopes.Elements())
}
