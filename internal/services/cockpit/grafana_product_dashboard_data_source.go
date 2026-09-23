package cockpit

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*GrafanaProductDashboardDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*GrafanaProductDashboardDataSource)(nil)
)

func NewGrafanaProductDashboardDataSource() datasource.DataSource {
	return &GrafanaProductDashboardDataSource{}
}

type GrafanaProductDashboardDataSource struct {
	globalAPI *cockpit.GlobalAPI
	meta      *meta.Meta
}

type grafanaProductDashboardDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	DashboardName types.String `tfsdk:"dashboard_name"`
	Name          types.String `tfsdk:"name"`
	Title         types.String `tfsdk:"title"`
	URL           types.String `tfsdk:"url"`
	Tags          types.List   `tfsdk:"tags"`
	Variables     types.List   `tfsdk:"variables"`
}

func (d *GrafanaProductDashboardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_grafana_product_dashboard"
}

//go:embed descriptions/grafana_product_dashboard_data_source.md
var grafanaProductDashboardDataSourceDescription string

func (d *GrafanaProductDashboardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: grafanaProductDashboardDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the project the dashboard belongs to",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"dashboard_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Grafana product dashboard to retrieve",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the data source (`{project_id}/{name}`)",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Dashboard name",
			},
			"title": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Human-readable dashboard title",
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URL to open the dashboard in Grafana",
			},
			"tags": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Dashboard tags",
			},
			"variables": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Dashboard variables",
			},
		},
	}
}

func (d *GrafanaProductDashboardDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
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

	d.meta = m
	d.globalAPI = globalAPI
}

func (d *GrafanaProductDashboardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state grafanaProductDashboardDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if d.globalAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Cockpit Global API",
			"The data source was not properly configured. Please report this issue to the provider developers.",
		)

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(state.ProjectID, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project_id", err.Error())

		return
	}

	dashboardName := state.DashboardName.ValueString()

	dashboard, err := retryOn403Value(ctx, func() (*cockpit.GrafanaProductDashboard, error) {
		return d.globalAPI.GetGrafanaProductDashboard(&cockpit.GlobalAPIGetGrafanaProductDashboardRequest{
			ProjectID:     projectID,
			DashboardName: dashboardName,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		if httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"Grafana product dashboard not found",
				fmt.Sprintf("Grafana product dashboard %q not found for project %s", dashboardName, projectID),
			)

			return
		}

		resp.Diagnostics.AddError(
			"Failed to get Grafana product dashboard",
			fmt.Sprintf("Could not retrieve dashboard %q for project %s: %s", dashboardName, projectID, err),
		)

		return
	}

	state.ID = types.StringValue(projectID + "/" + dashboard.Name)
	state.ProjectID = types.StringValue(projectID)
	state.DashboardName = types.StringValue(dashboard.Name)
	state.Name = types.StringValue(dashboard.Name)
	state.Title = types.StringValue(dashboard.Title)
	state.URL = types.StringValue(dashboard.URL)
	state.Tags = flattenStringList(ctx, dashboard.Tags, &resp.Diagnostics)
	state.Variables = flattenStringList(ctx, dashboard.Variables, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenStringList(ctx context.Context, values []string, diags *diag.Diagnostics) types.List {
	if values == nil {
		values = []string{}
	}

	list, d := types.ListValueFrom(ctx, types.StringType, values)
	diags.Append(d...)

	return list
}

func grafanaProductDashboardAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":      types.StringType,
		"title":     types.StringType,
		"url":       types.StringType,
		"tags":      types.ListType{ElemType: types.StringType},
		"variables": types.ListType{ElemType: types.StringType},
	}
}
