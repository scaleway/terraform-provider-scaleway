package cockpit

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*GrafanaSyncDataSourcesAction)(nil)
	_ action.ActionWithConfigure = (*GrafanaSyncDataSourcesAction)(nil)
)

type GrafanaSyncDataSourcesAction struct {
	globalAPI *cockpit.GlobalAPI
	meta      *meta.Meta
}

func (a *GrafanaSyncDataSourcesAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

	globalAPI, err := NewGlobalAPI(m)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring Cockpit Global API",
			fmt.Sprintf("Failed to configure the Cockpit Global API: %s", err),
		)

		return
	}

	a.globalAPI = globalAPI
	a.meta = m
}

func (a *GrafanaSyncDataSourcesAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_grafana_sync_data_sources"
}

type GrafanaSyncDataSourcesActionModel struct {
	ProjectID types.String `tfsdk:"project_id"`
}

func NewGrafanaSyncDataSourcesAction() action.Action {
	return &GrafanaSyncDataSourcesAction{}
}

//go:embed descriptions/grafanaSyncDataSources_action.md
var grafanaSyncDataSourcesActionDescription string

func (a *GrafanaSyncDataSourcesAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: grafanaSyncDataSourcesActionDescription,
		Description:         grafanaSyncDataSourcesActionDescription,
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the Project",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
		},
	}
}

func (a *GrafanaSyncDataSourcesAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data GrafanaSyncDataSourcesActionModel

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
			"The project_id attribute is required to synchronize Grafana data sources.",
		)

		return
	}

	err := a.globalAPI.SyncGrafanaDataSources(&cockpit.GlobalAPISyncGrafanaDataSourcesRequest{
		ProjectID: data.ProjectID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Cockpit SyncGrafanaDataSources action",
			fmt.Sprintf("Failed to synchronize Grafana data sources for project %s: %s", data.ProjectID.ValueString(), err),
		)

		return
	}
}
