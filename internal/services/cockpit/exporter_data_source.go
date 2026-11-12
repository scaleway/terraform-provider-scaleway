package cockpit

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

//go:embed descriptions/cockpit_exporter_data_source.md
var cockpitExporterDataSourceDescription string

var (
	_ datasource.DataSource              = (*ExporterDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ExporterDataSource)(nil)
)

func NewExporterDataSource() datasource.DataSource {
	return &ExporterDataSource{}
}

type ExporterDataSource struct {
	dataSourceBase
}

type exporterDataSourceModel struct {
	ExportedProducts   types.List             `tfsdk:"exported_products"`
	ID                 types.String           `tfsdk:"id"`
	Name               types.String           `tfsdk:"name"`
	Description        types.String           `tfsdk:"description"`
	DatasourceID       types.String           `tfsdk:"datasource_id"`
	Status             types.String           `tfsdk:"status"`
	CreatedAt          types.String           `tfsdk:"created_at"`
	UpdatedAt          types.String           `tfsdk:"updated_at"`
	SRN                types.String           `tfsdk:"srn"`
	ProjectID          types.String           `tfsdk:"project_id"`
	Region             types.String           `tfsdk:"region"`
	DatadogDestination []exporterDatadogModel `tfsdk:"datadog_destination"`
	OtlpDestination    []exporterOTLPModel    `tfsdk:"otlp_destination"`
}

func (d *ExporterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_exporter"
}

func (d *ExporterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitExporterDataSourceDescription,
		Description:         cockpitExporterDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id":                schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "ID of the exporter."},
			"name":              schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Name of the data export"},
			"description":       schema.StringAttribute{Computed: true, MarkdownDescription: "Description of the data export"},
			"datasource_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the data source linked to the data export"},
			"exported_products": schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "List of Scaleway products to export."},
			"status":            schema.StringAttribute{Computed: true, MarkdownDescription: "Status of the data export"},
			"created_at":        schema.StringAttribute{Computed: true, MarkdownDescription: "Date and time of the creation of the data export (RFC 3339 format)"},
			"updated_at":        schema.StringAttribute{Computed: true, MarkdownDescription: "Date and time of the last update of the data export (RFC 3339 format)"},
			"srn":               schema.StringAttribute{Computed: true, MarkdownDescription: "The Scaleway Resource Name (SRN) of the data export"},
			"project_id":        account.ProjectIDSchemaAttributeOptional("The project_id you want to attach the resource to"),
			"region":            regional.SchemaAttributeOptionalComputed(),
			"datadog_destination": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Datadog destination configuration",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"api_key":  schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Datadog API key"},
						"endpoint": schema.StringAttribute{Computed: true, MarkdownDescription: "Datadog endpoint URL"},
					},
				},
			},
			"otlp_destination": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "OTLP destination configuration",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"endpoint": schema.StringAttribute{Computed: true, MarkdownDescription: "OTLP endpoint URL"},
						"headers":  schema.MapAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Headers to include in requests"},
					},
				},
			},
		},
	}
}

func (d *ExporterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data exporterDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		api *cockpit.RegionalAPI
		exp *cockpit.Exporter
		err error
	)

	if data.ID.ValueString() != "" {
		api, exp, err = d.fetchByID(ctx, data.ID.ValueString())
	} else {
		api, exp, err = d.fetchByFilters(ctx, data)
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to read Cockpit exporter", err.Error())

		return
	}

	products, diags := types.ListValueFrom(ctx, types.StringType, exp.ExportedProducts)
	resp.Diagnostics.Append(diags...)

	datadog, otlp, diags := flattenExporterDestinations(ctx, exp, nil, nil)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(regional.NewIDString(exp.Region, exp.ID))
	data.Name = types.StringValue(exp.Name)
	data.Description = types.StringValue(exp.Description)
	data.DatasourceID = types.StringValue(regional.NewIDString(exp.Region, exp.DatasourceID))
	data.Status = types.StringValue(exp.Status.String())
	data.Region = types.StringValue(exp.Region.String())
	data.CreatedAt = flattenCockpitTime(exp.CreatedAt)
	data.UpdatedAt = flattenCockpitTime(exp.UpdatedAt)
	data.SRN = types.StringValue(exp.Srn)
	data.ExportedProducts = products
	data.DatadogDestination = datadog
	data.OtlpDestination = otlp

	if projectID, err := exporterProjectID(ctx, api, exp); err == nil {
		data.ProjectID = types.StringValue(projectID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *ExporterDataSource) fetchByID(ctx context.Context, regionalID string) (*cockpit.RegionalAPI, *cockpit.Exporter, error) {
	region, id, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, nil, err
	}

	api := cockpit.NewRegionalAPI(d.meta.ScwClient())

	exp, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.Exporter, error) {
		return api.GetExporter(&cockpit.RegionalAPIGetExporterRequest{
			Region:     region,
			ExporterID: id,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		return nil, nil, err
	}

	exp.Region = region // the API response may omit the region

	return api, exp, nil
}

func (d *ExporterDataSource) fetchByFilters(ctx context.Context, data exporterDataSourceModel) (*cockpit.RegionalAPI, *cockpit.Exporter, error) {
	api, region, err := d.regionalAPI(data.Region)
	if err != nil {
		return nil, nil, err
	}

	if data.ProjectID.ValueString() == "" {
		return nil, nil, errors.New("project_id is required when not fetching by id")
	}

	listReq := &cockpit.RegionalAPIListExportersRequest{
		Region:    region,
		ProjectID: data.ProjectID.ValueString(),
	}

	res, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.ListExportersResponse, error) {
		return api.ListExporters(listReq, scw.WithContext(ctx), scw.WithAllPages())
	})
	if err != nil {
		return nil, nil, err
	}

	if res.TotalCount == 0 {
		return nil, nil, errors.New("no exporter found matching the specified criteria")
	}

	for _, exp := range res.Exporters {
		if data.Name.ValueString() == "" || exp.Name == data.Name.ValueString() {
			exp.Region = region // the API response may omit the region

			return api, exp, nil
		}
	}

	return nil, nil, fmt.Errorf("no exporter found with name '%s'", data.Name.ValueString())
}
