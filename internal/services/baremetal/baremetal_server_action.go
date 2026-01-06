package baremetal

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*ServerBaremetalAction)(nil)
	_ action.ActionWithConfigure = (*ServerBaremetalAction)(nil)
)

type ServerBaremetalAction struct {
	baremetalAPI *baremetal.API
}

func (s *ServerBaremetalAction) Configure(ctx context.Context, request action.ConfigureRequest, response *action.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	m, ok := request.ProviderData.(*meta.Meta)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	s.baremetalAPI = baremetal.NewAPI(m.ScwClient())
}

func (s *ServerBaremetalAction) Metadata(ctx context.Context, request action.MetadataRequest, response *action.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_baremetal_server_action"
}

// ServerActionBaremetalModel defines the data structure for the action
type ServerActionBaremetalModel struct {
	ServerID types.String `tfsdk:"server_id"`
	Zone     types.String `tfsdk:"zone"`
	Action   types.String `tfsdk:"action"`
	BootType types.String `tfsdk:"boot_type"`
	Wait     types.Bool   `tfsdk:"wait"`
}

func NewBaremetalServerAction() action.Action {
	return &ServerBaremetalAction{}
}

//go:embed descriptions/server_action.md
var serverActionDescription string

func (s *ServerBaremetalAction) Schema(ctx context.Context, request action.SchemaRequest, response *action.SchemaResponse) {
	actionValues := []string{"reboot", "start", "stop"}

	response.Schema = schema.Schema{
		MarkdownDescription: serverActionDescription,
		Description:         serverActionDescription,
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Required:    true,
				Description: "Type of action to perform",
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(actionValues...),
				},
			},
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "Server id to send the action to",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"boot_type": schema.StringAttribute{
				Optional:    true,
				Description: "Boot type to use when rebooting the server. By default, set to `normal`",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(baremetal.ServerBootTypeNormal),
						string(baremetal.ServerBootTypeRescue),
					),
				},
			},
			"zone": zonal.SchemaAttribute("Zone of server to send the action to"),
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for server to finish action",
			},
		},
	}
}

func (s *ServerBaremetalAction) Invoke(ctx context.Context, request action.InvokeRequest, response *action.InvokeResponse) {
	var data ServerActionBaremetalModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	if s.baremetalAPI == nil {
		response.Diagnostics.AddError(
			"Unconfigured baremetalAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	zone, serverID, _ := locality.ParseLocalizedID(data.ServerID.ValueString())
	if zone == "" {
		if !data.Zone.IsNull() {
			zone = data.Zone.ValueString()
		} else {
			response.Diagnostics.AddError(
				"missing zone in config",
				fmt.Sprintf("zone could not be extracted from either the action configuration or the resource ID (%s)",
					data.ServerID.ValueString(),
				),
			)

			return
		}
	}

	switch data.Action.ValueString() {
	case "start":
		actionReq := &baremetal.StartServerRequest{
			Zone:     scw.Zone(zone),
			ServerID: serverID,
		}

		_, err := s.baremetalAPI.StartServer(actionReq)
		if err != nil {
			response.Diagnostics.AddError(
				"failed to start server",
				err.Error(),
			)

			return
		}
	case "stop":
		actionReq := &baremetal.StopServerRequest{
			Zone:     scw.Zone(zone),
			ServerID: serverID,
		}

		_, err := s.baremetalAPI.StopServer(actionReq)
		if err != nil {
			response.Diagnostics.AddError(
				"failed to stop server",
				err.Error(),
			)

			return
		}
	case "reboot":
		bootType := baremetal.ServerBootType("normal") // default boot type
		if !data.BootType.IsNull() && !data.BootType.IsUnknown() {
			bootType = baremetal.ServerBootType(data.BootType.ValueString())
		}

		actionReq := &baremetal.RebootServerRequest{
			Zone:     scw.Zone(zone),
			ServerID: serverID,
			BootType: bootType,
		}

		_, err := s.baremetalAPI.RebootServer(actionReq)
		if err != nil {
			response.Diagnostics.AddError(
				"failed to reboot server",
				err.Error(),
			)

			return
		}
	}

	if data.Wait.ValueBool() {
		_, err := s.baremetalAPI.WaitForServer(&baremetal.WaitForServerRequest{
			Zone:     scw.Zone(zone),
			ServerID: serverID,
		}, scw.WithContext(ctx))
		if err != nil {
			response.Diagnostics.AddError(
				"error waiting for server"+serverID,
				err.Error())
		}
	}
}
