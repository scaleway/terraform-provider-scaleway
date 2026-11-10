package cockpit

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

//go:embed descriptions/cockpit_grafana_data_source.md
var cockpitGrafanaDataSourceDescription string

var (
	_ datasource.DataSource              = (*GrafanaDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*GrafanaDataSource)(nil)
)

func NewGrafanaDataSource() datasource.DataSource {
	return &GrafanaDataSource{}
}

type GrafanaDataSource struct {
	dataSourceBase
}

type grafanaDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	GrafanaURL types.String `tfsdk:"grafana_url"`
	SRN        types.String `tfsdk:"srn"`
}

func (d *GrafanaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_grafana"
}

func (d *GrafanaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitGrafanaDataSourceDescription,
		Description:         cockpitGrafanaDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "The project ID"},
			"project_id":  account.ProjectIDSchemaAttributeOptional("The project ID associated with the Grafana instance"),
			"grafana_url": schema.StringAttribute{Computed: true, MarkdownDescription: "The URL to access the Grafana dashboard"},
			"srn":         schema.StringAttribute{Computed: true, MarkdownDescription: "The Scaleway Resource Name (SRN) of the Grafana instance"},
		},
	}
}

func (d *GrafanaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data grafanaDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project_id", err.Error())

		return
	}

	api := cockpit.NewGlobalAPI(d.meta.ScwClient())

	grafana, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.Grafana, error) {
		return api.GetGrafana(&cockpit.GlobalAPIGetGrafanaRequest{
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		if httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"Grafana instance not found",
				fmt.Sprintf("Grafana instance not found for project %s. Ensure that Cockpit is activated for this project.", projectID),
			)

			return
		}

		resp.Diagnostics.AddError("Failed to read Grafana instance", err.Error())

		return
	}

	data.ID = types.StringValue(projectID)
	data.ProjectID = types.StringValue(projectID)
	data.GrafanaURL = types.StringValue(grafana.GrafanaURL)
	data.SRN = types.StringValue(grafana.Srn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
