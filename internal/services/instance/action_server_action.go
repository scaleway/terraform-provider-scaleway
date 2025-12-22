package instance

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*ServerAction)(nil)
	_ action.ActionWithConfigure = (*ServerAction)(nil)
)

type ServerAction struct {
	instanceAPI *instance.API
}

func (a *ServerAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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
	a.instanceAPI = instance.NewAPI(client)
}

func (a *ServerAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_server_action"
}

type ServerActionModel struct {
	ServerID types.String `tfsdk:"server_id"`
	Zone     types.String `tfsdk:"zone"`
	Action   types.String `tfsdk:"action"`
	Wait     types.Bool   `tfsdk:"wait"`
}

func NewServerAction() action.Action {
	return &ServerAction{}
}

//go:embed descriptions/server_action_action.md
var serverActionDescription string

func (a *ServerAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	actionsValues := instance.ServerAction("").Values()

	actionStringValues := make([]string, 0, len(actionsValues))
	for _, actionValue := range actionsValues {
		actionStringValues = append(actionStringValues, actionValue.String())
	}

	resp.Schema = schema.Schema{
		Description:         serverActionDescription,
		MarkdownDescription: serverActionDescription,
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Required:    true,
				Description: "Type of action to perform",
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(actionStringValues...),
				},
			},
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

func (a *ServerAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerActionModel
	// Read action config data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.instanceAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured instanceAPI",
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
			resp.Diagnostics.AddError(
				"missing zone in config",
				fmt.Sprintf("zone could not be extracted from either the action configuration or the resource ID (%s)",
					data.ServerID.ValueString(),
				),
			)

			return
		}
	}

	actionReq := &instance.ServerActionRequest{
		ServerID: serverID,
		Zone:     scw.Zone(zone),
		Action:   instance.ServerAction(data.Action.ValueString()),
	}

	_, err := a.instanceAPI.ServerAction(actionReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"error in server action",
			err.Error())

		return
	}

	if data.Wait.ValueBool() {
		server, err := a.instanceAPI.WaitForServer(&instance.WaitForServerRequest{
			ServerID: serverID,
			Zone:     scw.Zone(zone),
		}, scw.WithContext(ctx))
		if err != nil && data.Action.ValueString() != instance.ServerActionTerminate.String() {
			resp.Diagnostics.AddError(
				"error waiting for server"+serverID,
				err.Error())
		}

		if data.Action.ValueString() == instance.ServerActionBackup.String() && server != nil {
			for _, volume := range server.Volumes {
				_, err := a.instanceAPI.WaitForVolume(&instance.WaitForVolumeRequest{
					VolumeID: volume.ID,
					Zone:     scw.Zone(zone),
				}, scw.WithContext(ctx))
				if err != nil {
					resp.Diagnostics.AddError(
						"error waiting for volume "+volume.ID,
						err.Error())
				}
			}
		}
	}
}
