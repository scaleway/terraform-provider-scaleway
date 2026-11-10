package cockpit

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

//go:embed descriptions/cockpit_config_data_source.md
var cockpitConfigDataSourceDescription string

var (
	_ datasource.DataSource              = (*ConfigDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ConfigDataSource)(nil)
)

func NewConfigDataSource() datasource.DataSource {
	return &ConfigDataSource{}
}

type ConfigDataSource struct {
	dataSourceBase
}

type retentionModel struct {
	MinDays     types.Int64 `tfsdk:"min_days"`
	MaxDays     types.Int64 `tfsdk:"max_days"`
	DefaultDays types.Int64 `tfsdk:"default_days"`
}

type configDataSourceModel struct {
	ID                      types.String     `tfsdk:"id"`
	Region                  types.String     `tfsdk:"region"`
	CustomMetricsRetention  []retentionModel `tfsdk:"custom_metrics_retention"`
	CustomLogsRetention     []retentionModel `tfsdk:"custom_logs_retention"`
	CustomTracesRetention   []retentionModel `tfsdk:"custom_traces_retention"`
	ProductMetricsRetention []retentionModel `tfsdk:"product_metrics_retention"`
	ProductLogsRetention    []retentionModel `tfsdk:"product_logs_retention"`
}

func (d *ConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_config"
}

func (d *ConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitConfigDataSourceDescription,
		Description:         cockpitConfigDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id":                        schema.StringAttribute{Computed: true, MarkdownDescription: "The region of the configuration."},
			"region":                    regional.SchemaAttributeOptionalComputed(),
			"custom_metrics_retention":  retentionAttribute("Retention limits and default for custom metrics data sources."),
			"custom_logs_retention":     retentionAttribute("Retention limits and default for custom logs data sources."),
			"custom_traces_retention":   retentionAttribute("Retention limits and default for custom traces data sources."),
			"product_metrics_retention": retentionAttribute("Retention limits and default for Scaleway product metrics data sources."),
			"product_logs_retention":    retentionAttribute("Retention limits and default for Scaleway product logs data sources."),
		},
	}
}

func retentionAttribute(description string) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed:            true,
		MarkdownDescription: description,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"min_days":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Minimum retention in days."},
				"max_days":     schema.Int64Attribute{Computed: true, MarkdownDescription: "Maximum retention in days."},
				"default_days": schema.Int64Attribute{Computed: true, MarkdownDescription: "Default retention in days."},
			},
		},
	}
}

func (d *ConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data configDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	api, region, err := d.regionalAPI(data.Region)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	res, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.GetConfigResponse, error) {
		return api.GetConfig(&cockpit.RegionalAPIGetConfigRequest{
			Region: region,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Cockpit config", err.Error())

		return
	}

	data.ID = types.StringValue(region.String())
	data.Region = types.StringValue(region.String())
	data.CustomMetricsRetention = flattenRetention(res.CustomMetricsRetention)
	data.CustomLogsRetention = flattenRetention(res.CustomLogsRetention)
	data.CustomTracesRetention = flattenRetention(res.CustomTracesRetention)
	data.ProductMetricsRetention = flattenRetention(res.ProductMetricsRetention)
	data.ProductLogsRetention = flattenRetention(res.ProductLogsRetention)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenRetention(retention *cockpit.GetConfigResponseRetention) []retentionModel {
	if retention == nil {
		return nil
	}

	return []retentionModel{{
		MinDays:     types.Int64Value(int64(retention.MinDays)),
		MaxDays:     types.Int64Value(int64(retention.MaxDays)),
		DefaultDays: types.Int64Value(int64(retention.DefaultDays)),
	}}
}
