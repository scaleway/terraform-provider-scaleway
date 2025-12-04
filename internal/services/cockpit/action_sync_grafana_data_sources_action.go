package cockpit

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*SyncGrafanaDataSourcesAction)(nil)
	_ action.ActionWithConfigure = (*SyncGrafanaDataSourcesAction)(nil)
)

type SyncGrafanaDataSourcesAction struct {
	globalAPI *cockpit.GlobalAPI
}

func (a *SyncGrafanaDataSourcesAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

	client := m.ScwClient()
	a.globalAPI = cockpit.NewGlobalAPI(client)
}

func (a *SyncGrafanaDataSourcesAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_sync_grafana_data_sources_action"
}

type SyncGrafanaDataSourcesActionModel struct {
	ProjectID types.String `tfsdk:"project_id"`
}

func NewSyncGrafanaDataSourcesAction() action.Action {
	return &SyncGrafanaDataSourcesAction{}
}

func (a *SyncGrafanaDataSourcesAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the Project",
			},
		},
	}
}

func (a *SyncGrafanaDataSourcesAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data SyncGrafanaDataSourcesActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.globalAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured globalAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.ProjectID.IsNull() || data.ProjectID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing project_id",
			"The project_id attribute is required to sync Grafana data sources.",
		)

		return
	}

	err := a.globalAPI.SyncGrafanaDataSources(&cockpit.GlobalAPISyncGrafanaDataSourcesRequest{
		ProjectID: data.ProjectID.ValueString(),
	}, scw.WithContext(ctx))

	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Cockpit SyncGrafanaDataSources action",
			fmt.Sprintf("Failed to sync Grafana data sources for project %s: %s", data.ProjectID.ValueString(), err),
		)

		return
	}
}

