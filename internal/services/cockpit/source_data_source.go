package cockpit

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

//go:embed descriptions/cockpit_source_data_source.md
var cockpitSourceDataSourceDescription string

var (
	_ datasource.DataSource              = (*SourceDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*SourceDataSource)(nil)
)

func NewSourceDataSource() datasource.DataSource {
	return &SourceDataSource{}
}

type SourceDataSource struct {
	dataSourceBase
}

// sourceDataSourceModel is shared by the source data source and the elements of the sources data source.
type sourceDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Type                    types.String `tfsdk:"type"`
	URL                     types.String `tfsdk:"url"`
	Origin                  types.String `tfsdk:"origin"`
	Region                  types.String `tfsdk:"region"`
	ProjectID               types.String `tfsdk:"project_id"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
	PushURL                 types.String `tfsdk:"push_url"`
	SRN                     types.String `tfsdk:"srn"`
	RetentionDays           types.Int64  `tfsdk:"retention_days"`
	SynchronizedWithGrafana types.Bool   `tfsdk:"synchronized_with_grafana"`
}

func (d *SourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_source"
}

// sourceDataSourceComputedAttributes returns the computed attributes common to the source and sources data sources.
func sourceDataSourceComputedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url":                       schema.StringAttribute{Computed: true, MarkdownDescription: "The URL of the datasource"},
		"synchronized_with_grafana": schema.BoolAttribute{Computed: true, MarkdownDescription: "Indicates whether the data source is synchronized with Grafana"},
		"retention_days":            schema.Int64Attribute{Computed: true, MarkdownDescription: "The number of days to retain data. Use scaleway_cockpit_config data source to read allowed min, max, and default values for each data source type."},
		"created_at":                schema.StringAttribute{Computed: true, MarkdownDescription: "The date and time of the creation of the cockpit datasource"},
		"updated_at":                schema.StringAttribute{Computed: true, MarkdownDescription: "The date and time of the last update of the cockpit datasource"},
		"push_url":                  schema.StringAttribute{Computed: true, MarkdownDescription: "The URL endpoint used for pushing data to the cockpit data source."},
		"srn":                       schema.StringAttribute{Computed: true, MarkdownDescription: "The Scaleway Resource Name (SRN) of the data source"},
	}
}

func (d *SourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := sourceDataSourceComputedAttributes()
	attributes["id"] = schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "ID of the data source."}
	attributes["name"] = schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Name of the datasource"}
	attributes["type"] = schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "The type of the datasource"}
	attributes["origin"] = schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "The origin of the datasource"}
	attributes["region"] = dataSourceRegionAttribute()
	attributes["project_id"] = dataSourceProjectIDAttribute("The project_id you want to attach the resource to")

	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitSourceDataSourceDescription,
		Description:         cockpitSourceDataSourceDescription,
		Attributes:          attributes,
	}
}

func (d *SourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sourceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		ds  *cockpit.DataSource
		err error
	)

	if data.ID.ValueString() != "" {
		ds, err = d.fetchByID(ctx, data.ID.ValueString())
	} else {
		ds, err = d.fetchByFilters(ctx, data)
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to read Cockpit source", err.Error())

		return
	}

	state := flattenSourceDataSource(ds)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *SourceDataSource) fetchByID(ctx context.Context, regionalID string) (*cockpit.DataSource, error) {
	region, id, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, err
	}

	return transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.DataSource, error) {
		return cockpit.NewRegionalAPI(d.meta.ScwClient()).GetDataSource(&cockpit.RegionalAPIGetDataSourceRequest{
			Region:       region,
			DataSourceID: id,
		}, scw.WithContext(ctx))
	})
}

func (d *SourceDataSource) fetchByFilters(ctx context.Context, data sourceDataSourceModel) (*cockpit.DataSource, error) {
	api, region, err := d.regionalAPI(data.Region)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if res.TotalCount == 0 {
		return nil, errors.New("no data source found matching the specified criteria")
	}

	if data.Name.ValueString() == "" {
		return res.DataSources[0], nil
	}

	for _, ds := range res.DataSources {
		if ds.Name == data.Name.ValueString() {
			return ds, nil
		}
	}

	return nil, fmt.Errorf("no data source found with name '%s'", data.Name.ValueString())
}

// flattenSourceDataSource builds the data source model; the push URL is left empty for an unknown source type.
func flattenSourceDataSource(ds *cockpit.DataSource) sourceDataSourceModel {
	pushURL, _ := createCockpitPushURL(ds.Type, ds.URL)

	return sourceDataSourceModel{
		ID:                      types.StringValue(regional.NewIDString(ds.Region, ds.ID)),
		Name:                    types.StringValue(ds.Name),
		Type:                    types.StringValue(ds.Type.String()),
		URL:                     types.StringValue(ds.URL),
		Origin:                  types.StringValue(ds.Origin.String()),
		Region:                  types.StringValue(ds.Region.String()),
		ProjectID:               types.StringValue(ds.ProjectID),
		CreatedAt:               flattenCockpitTime(ds.CreatedAt),
		UpdatedAt:               flattenCockpitTime(ds.UpdatedAt),
		PushURL:                 types.StringValue(pushURL),
		SRN:                     types.StringValue(ds.Srn),
		RetentionDays:           types.Int64Value(int64(ds.RetentionDays)),
		SynchronizedWithGrafana: types.BoolValue(ds.SynchronizedWithGrafana),
	}
}

// flattenCockpitTime formats an optional API time as RFC 3339, null when absent.
func flattenCockpitTime(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}

	return types.StringValue(t.Format(time.RFC3339))
}
