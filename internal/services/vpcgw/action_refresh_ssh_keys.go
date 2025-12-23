package vpcgw

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*RefreshSSHKeysAction)(nil)
	_ action.ActionWithConfigure = (*RefreshSSHKeysAction)(nil)
)

type RefreshSSHKeysAction struct {
	vpcgwAPI *vpcgw.API
	meta     *meta.Meta
}

func (a *RefreshSSHKeysAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *scw.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client := m.ScwClient()
	a.vpcgwAPI = vpcgw.NewAPI(client)
	a.meta = m
}

func (a *RefreshSSHKeysAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcgw_refresh_ssh_keys_action"
}

type RefreshSSHKeysActionModel struct {
	GatewayID types.String `tfsdk:"gateway_id"`
	Zone      types.String `tfsdk:"zone"`
	Wait      types.Bool   `tfsdk:"wait"`
}

func NewRefreshSSHKeysAction() action.Action {
	return &RefreshSSHKeysAction{}
}

func (a *RefreshSSHKeysAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"gateway_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the gateway to refresh SSH keys on (UUID format)",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: "Zone of the gateway. If not set, the zone is derived from the gateway_id when possible or from the provider configuration",
			},
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for the SSH keys refresh to complete before returning",
			},
		},
	}
}

func (a *RefreshSSHKeysAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RefreshSSHKeysActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.vpcgwAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured vpcgwAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.GatewayID.IsNull() || data.GatewayID.IsUnknown() || data.GatewayID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing gateway_id",
			"The gateway_id attribute is required to refresh the ssh keys.",
		)

		return
	}

	gatewayID := locality.ExpandID(data.GatewayID.ValueString())

	var zone scw.Zone

	if !data.Zone.IsNull() && !data.Zone.IsUnknown() && data.Zone.ValueString() != "" {
		parsedZone, err := scw.ParseZone(data.Zone.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid zone value",
				fmt.Sprintf("The zone attribute must be a valid Scaleway zone. Got %q: %s", data.Zone.ValueString(), err),
			)

			return
		}

		zone = parsedZone
	} else {
		// Try to derive zone from the gateway_id if it is a zonal ID.
		if derivedZone, id, parseErr := zonal.ParseID(data.GatewayID.ValueString()); parseErr == nil {
			zone = derivedZone
			gatewayID = id
		} else if a.meta != nil {
			// Use default zone from provider configuration
			defaultZone, exists := a.meta.ScwClient().GetDefaultZone()
			if !exists {
				resp.Diagnostics.AddError(
					"Missing zone",
					"The zone attribute is required to refresh ssh keys. Please provide it explicitly or configure a default zone in the provider.",
				)

				return
			}

			zone = defaultZone
		}
	}

	refreshKeysReq := &vpcgw.RefreshSSHKeysRequest{
		Zone:      zone,
		GatewayID: gatewayID,
	}

	_, err := a.vpcgwAPI.RefreshSSHKeys(refreshKeysReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing vpcgw RefreshSSHKeys action", fmt.Sprintf("%s", err),
		)

		return
	}

	if data.Wait.ValueBool() {
		_, err = waitForVPCPublicGatewayV2(ctx, a.vpcgwAPI, zone, gatewayID, defaultTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for vpcgw SSH keys refresh completion",
				fmt.Sprintf("SSH keys refresh for gateway %s did not complete: %s", gatewayID, err),
			)

			return
		}
	}
}
