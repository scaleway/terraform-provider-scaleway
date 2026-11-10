package cockpit

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeTokenStateV0(t *testing.T) {
	ctx := t.Context()

	schemaResp := &resource.SchemaResponse{}
	(&TokenResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)

	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(`{"id":"3ed23100-398f-4b39-b3ab-fadd1b693da5","name":"tok","secret_key":"secret"}`)},
	}
	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}

	upgradeTokenStateV0(ctx, req, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	var state tokenResourceModel

	require.False(t, resp.State.Get(ctx, &state).HasError())
	assert.Equal(t, "fr-par/3ed23100-398f-4b39-b3ab-fadd1b693da5", state.ID.ValueString())
	assert.Equal(t, "fr-par", state.Region.ValueString())
	assert.Equal(t, "secret", state.SecretKey.ValueString())
	assert.True(t, state.Scopes.IsNull())
}
