package s2svpn

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*ConnectionDisableRoutePropagationAction)(nil)
	_ action.ActionWithConfigure = (*ConnectionDisableRoutePropagationAction)(nil)
)

type ConnectionDisableRoutePropagationAction struct {
	s2svpnAPI *s2s_vpn.API
	meta      *meta.Meta
}

func (a *ConnectionDisableRoutePropagationAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.meta = m
	a.s2svpnAPI = s2s_vpn.NewAPI(m.ScwClient())
}

func (a *ConnectionDisableRoutePropagationAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s2s_vpn_connection_disable_route_propagation"
}

type ConnectionDisableRoutePropagationActionModel struct {
	ConnectionID types.String `tfsdk:"connection_id"`
	Region       types.String `tfsdk:"region"`
}

func NewConnectionDisableRoutePropagationAction() action.Action {
	return &ConnectionDisableRoutePropagationAction{}
}

//go:embed descriptions/connection_disable_route_propagation_action.md
var connectionDisableRoutePropagationActionDescription string

func (a *ConnectionDisableRoutePropagationAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         connectionDisableRoutePropagationActionDescription,
		MarkdownDescription: connectionDisableRoutePropagationActionDescription,
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{
				Required:    true,
				Description: "The S2S VPN connection ID on which to disable route propagation. Can be a plain UUID or a regional ID (region/uuid).",
			},
			"region": regional.SchemaAttribute(),
		},
	}
}

func (a *ConnectionDisableRoutePropagationAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ConnectionDisableRoutePropagationActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.s2svpnAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured S2S VPN API",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.ConnectionID.IsNull() || data.ConnectionID.IsUnknown() || data.ConnectionID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing connection_id",
			"The connection_id attribute is required to disable route propagation on a connection.",
		)

		return
	}

	connectionID := locality.ExpandID(data.ConnectionID.ValueString())

	var region scw.Region

	if !data.Region.IsNull() && !data.Region.IsUnknown() && data.Region.ValueString() != "" {
		parsedRegion, err := scw.ParseRegion(data.Region.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid region value",
				fmt.Sprintf("The region attribute must be a valid Scaleway region. Got %q: %s", data.Region.ValueString(), err),
			)

			return
		}

		region = parsedRegion
	} else {
		if derivedRegion, id, parseErr := regional.ParseID(data.ConnectionID.ValueString()); parseErr == nil {
			region = derivedRegion
			connectionID = id
		} else if a.meta != nil {
			defaultRegion, exists := a.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Unable to determine region",
					"Failed to get default region from provider configuration. Please set the region attribute, use a regional connection_id (region/uuid), or configure a default region in the provider.",
				)

				return
			}

			region = defaultRegion
		}
	}

	if region == "" {
		resp.Diagnostics.AddError(
			"Missing region",
			"Could not determine region for S2S VPN connection. Please set the region attribute, use a regional connection_id (region/uuid), or configure a default region in the provider.",
		)

		return
	}

	connection, err := a.s2svpnAPI.GetConnection(&s2s_vpn.GetConnectionRequest{
		Region:       region,
		ConnectionID: connectionID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting S2S VPN connection",
			fmt.Sprintf("Failed to get connection %s: %s", connectionID, err),
		)

		return
	}

	_, err = waitForVPNGateway(ctx, a.s2svpnAPI, region, connection.VpnGatewayID, defaulVPNGatewayTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for VPN gateway",
			fmt.Sprintf("Failed to wait for VPN gateway %s to be ready: %s", connection.VpnGatewayID, err),
		)

		return
	}

	disableReq := &s2s_vpn.DisableRoutePropagationRequest{
		Region:       region,
		ConnectionID: connectionID,
	}

	_, err = a.s2svpnAPI.DisableRoutePropagation(disableReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing S2S VPN DisableRoutePropagation action",
			fmt.Sprintf("Failed to disable route propagation on connection %s: %s", connectionID, err),
		)

		return
	}
}
