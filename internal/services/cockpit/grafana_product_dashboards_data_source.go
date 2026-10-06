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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*GrafanaProductDashboardsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*GrafanaProductDashboardsDataSource)(nil)
)

func NewGrafanaProductDashboardsDataSource() datasource.DataSource {
	return &GrafanaProductDashboardsDataSource{}
}

type GrafanaProductDashboardsDataSource struct {
	globalAPI *cockpit.GlobalAPI
	meta      *meta.Meta
}

type grafanaProductDashboardsDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	Tags       types.List   `tfsdk:"tags"`
	Dashboards types.List   `tfsdk:"dashboards"`
	Names      types.List   `tfsdk:"names"`
}

func (d *GrafanaProductDashboardsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_grafana_product_dashboards"
}

//go:embed descriptions/grafana_product_dashboards_data_source.md
var grafanaProductDashboardsDataSourceDescription string

func (d *GrafanaProductDashboardsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: grafanaProductDashboardsDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the project to list Grafana product dashboards for",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter dashboards by tags (e.g. rdb, lb)",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The project ID",
			},
			"dashboards": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of Grafana product dashboards",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Dashboard name (use with scaleway_cockpit_grafana_product_dashboard)",
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
				},
			},
			"names": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of dashboard names",
			},
		},
	}
}

func (d *GrafanaProductDashboardsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GrafanaProductDashboardsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state grafanaProductDashboardsDataSourceModel

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

	listReq := &cockpit.GlobalAPIListGrafanaProductDashboardsRequest{
		ProjectID: projectID,
	}

	if !state.Tags.IsNull() && !state.Tags.IsUnknown() {
		var tags []string

		resp.Diagnostics.Append(state.Tags.ElementsAs(ctx, &tags, false)...)

		if resp.Diagnostics.HasError() {
			return
		}

		listReq.Tags = tags
	}

	apiResp, err := retryOn403Value(ctx, func() (*cockpit.ListGrafanaProductDashboardsResponse, error) {
		return d.globalAPI.ListGrafanaProductDashboards(listReq, scw.WithContext(ctx), scw.WithAllPages())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list Grafana product dashboards",
			fmt.Sprintf("Could not list dashboards for project %s: %s", projectID, err),
		)

		return
	}

	dashboards, names := flattenGrafanaProductDashboards(ctx, apiResp.Dashboards, &resp.Diagnostics)

	state.ID = types.StringValue(projectID)
	state.ProjectID = types.StringValue(projectID)
	state.Dashboards = dashboards
	state.Names = names

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenGrafanaProductDashboards(ctx context.Context, dashboards []*cockpit.GrafanaProductDashboard, diags *diag.Diagnostics) (types.List, types.List) {
	itemType := types.ObjectType{AttrTypes: grafanaProductDashboardAttrTypes()}
	items := make([]attr.Value, 0, len(dashboards))
	names := make([]string, 0, len(dashboards))

	for _, dashboard := range dashboards {
		if dashboard == nil {
			continue
		}

		obj, d := types.ObjectValue(grafanaProductDashboardAttrTypes(), map[string]attr.Value{
			"name":      types.StringValue(dashboard.Name),
			"title":     types.StringValue(dashboard.Title),
			"url":       types.StringValue(dashboard.URL),
			"tags":      flattenStringList(ctx, dashboard.Tags, diags),
			"variables": flattenStringList(ctx, dashboard.Variables, diags),
		})
		diags.Append(d...)

		items = append(items, obj)
		names = append(names, dashboard.Name)
	}

	dashboardsList, d := types.ListValue(itemType, items)
	diags.Append(d...)

	namesList := flattenStringList(ctx, names, diags)

	return dashboardsList, namesList
}
