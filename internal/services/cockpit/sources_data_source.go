package cockpit

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

//go:embed descriptions/cockpit_sources_data_source.md
var cockpitSourcesDataSourceDescription string

var (
	_ datasource.DataSource              = (*SourcesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*SourcesDataSource)(nil)
)

func NewSourcesDataSource() datasource.DataSource {
	return &SourcesDataSource{}
}

type SourcesDataSource struct {
	dataSourceBase
}

type sourcesDataSourceModel struct {
	ID        types.String            `tfsdk:"id"`
	Name      types.String            `tfsdk:"name"`
	Type      types.String            `tfsdk:"type"`
	Origin    types.String            `tfsdk:"origin"`
	Region    types.String            `tfsdk:"region"`
	ProjectID types.String            `tfsdk:"project_id"`
	Sources   []sourceDataSourceModel `tfsdk:"sources"`
}

func (d *SourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_sources"
}

func (d *SourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	element := sourceDataSourceComputedAttributes()
	element["id"] = schema.StringAttribute{Computed: true, MarkdownDescription: "The ID of the data source."}
	element["name"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Name of the datasource"}
	element["type"] = schema.StringAttribute{Computed: true, MarkdownDescription: "The type of the datasource"}
	element["origin"] = schema.StringAttribute{Computed: true, MarkdownDescription: "The origin of the datasource"}
	element["region"] = schema.StringAttribute{Computed: true, MarkdownDescription: "The region you want to attach the resource to"}
	element["project_id"] = schema.StringAttribute{Computed: true, MarkdownDescription: "The project_id you want to attach the resource to"}

	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitSourcesDataSourceDescription,
		Description:         cockpitSourcesDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "The region of the data sources."},
			"name":       schema.StringAttribute{Optional: true, MarkdownDescription: "Name of the datasource"},
			"type":       schema.StringAttribute{Optional: true, MarkdownDescription: "The type of the datasource"},
			"origin":     schema.StringAttribute{Optional: true, MarkdownDescription: "The origin of the datasource"},
			"region":     regional.SchemaAttributeOptionalComputed(),
			"project_id": account.ProjectIDSchemaAttributeOptional("The project_id you want to attach the resource to"),
			"sources": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of cockpit sources.",
				NestedObject:        schema.NestedAttributeObject{Attributes: element},
			},
		},
	}
}

func (d *SourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sourcesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	api, region, err := d.regionalAPI(data.Region)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	listReq := &cockpit.RegionalAPIListDataSourcesRequest{
		Region:    region,
		ProjectID: data.ProjectID.ValueString(),
	}

	if data.Type.ValueString() != "" {
		listReq.Types = []cockpit.DataSourceType{cockpit.DataSourceType(data.Type.ValueString())}
	}

	if data.Origin.ValueString() != "" {
		listReq.Origin = cockpit.DataSourceOrigin(data.Origin.ValueString())
	}

	res, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.ListDataSourcesResponse, error) {
		return api.ListDataSources(listReq, scw.WithContext(ctx), scw.WithAllPages())
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list Cockpit sources", err.Error())

		return
	}

	data.Sources = make([]sourceDataSourceModel, 0, len(res.DataSources))

	for _, ds := range res.DataSources {
		if data.Name.ValueString() != "" && ds.Name != data.Name.ValueString() {
			continue
		}

		data.Sources = append(data.Sources, flattenSourceDataSource(ds))
	}

	data.ID = types.StringValue(region.String())
	data.Region = types.StringValue(region.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
