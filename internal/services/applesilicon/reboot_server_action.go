package applesilicon

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*ServerAppleSiliconRebootAction)(nil)
	_ action.ActionWithConfigure = (*ServerAppleSiliconRebootAction)(nil)
)

type ServerAppleSiliconRebootAction struct {
	appleSiliconAPI *applesilicon.API
}

func (s *ServerAppleSiliconRebootAction) Configure(ctx context.Context, request action.ConfigureRequest, response *action.ConfigureResponse) {
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

	s.appleSiliconAPI = applesilicon.NewAPI(m.ScwClient())
}

func (s *ServerAppleSiliconRebootAction) Metadata(ctx context.Context, request action.MetadataRequest, response *action.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_apple_silicon_reboot_server_action"
}

// StartDiagnosticActionAppleSiliconModel defines the data structure for the action
type ServerActionAppleSiliconModel struct {
	ServerID types.String `tfsdk:"server_id"`
	Zone     types.String `tfsdk:"zone"`
	Wait     types.Bool   `tfsdk:"wait"`
}

func NewRebootServerAction() action.Action {
	return &ServerAppleSiliconRebootAction{}
}

func (s *ServerAppleSiliconRebootAction) Schema(ctx context.Context, request action.SchemaRequest, response *action.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "Server id to send the action to",
				Validators: []validator.String{
					verify.IsStringUUID(),
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

func (s *ServerAppleSiliconRebootAction) Invoke(ctx context.Context, request action.InvokeRequest, response *action.InvokeResponse) {
	var data ServerActionAppleSiliconModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	if s.appleSiliconAPI == nil {
		response.Diagnostics.AddError(
			"Unconfigured appleSiliconAPI",
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

	_, err := s.appleSiliconAPI.RebootServer(&applesilicon.RebootServerRequest{
		Zone:     scw.Zone(zone),
		ServerID: serverID,
	})
	if err != nil {
		response.Diagnostics.AddError(
			"failed to reboot server",
			err.Error(),
		)

		return
	}

	if data.Wait.ValueBool() {
		_, err = s.appleSiliconAPI.WaitForServer(&applesilicon.WaitForServerRequest{
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
