package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type ServerReboot struct {
	instanceAPI *instance.API
}

func (a *ServerReboot) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*scw.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *scw.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.instanceAPI = instance.NewAPI(client)
}

func (a *ServerReboot) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_server_reboot"
}

type ServerRebootModel struct {
	ServerID types.String `tfsdk:"server_id"`
	Zone     types.String `tfsdk:"zone"`
	Wait     types.Bool   `tfsdk:"wait"`
}

func NewServerReboot() action.Action {
	return &ServerReboot{}
}

func (a *ServerReboot) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "Server id to reboot",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: "Zone of server to reboot",
			},
			"wait": schema.BoolAttribute{
				Optional:    true,
				Description: "Wait for server to finish reboot",
			},
		},
	}
}

func (a *ServerReboot) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ServerRebootModel
	// Read action config data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := a.instanceAPI.ServerAction(&instance.ServerActionRequest{
		ServerID: data.ServerID.String(),
		Zone:     scw.Zone(data.Zone.String()),
		Action:   instance.ServerActionReboot,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"error in server action",
			fmt.Sprintf("%s", err))
	}
	if data.Wait.ValueBool() {
		_, errWait := a.instanceAPI.WaitForServer(&instance.WaitForServerRequest{
			ServerID: data.ServerID.String(),
			Zone:     scw.Zone(data.Zone.String()),
		})
		if errWait != nil {
			resp.Diagnostics.AddError(
				"error in wait server",
				fmt.Sprintf("%s", err))
		}
	}
}
