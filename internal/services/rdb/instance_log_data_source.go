package rdb

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	rdb "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*InstanceLogDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*InstanceLogDataSource)(nil)
)

func NewInstanceLogDataSource() datasource.DataSource {
	return &InstanceLogDataSource{}
}

type InstanceLogDataSource struct {
	rdbAPI *rdb.API
	meta   *meta.Meta
}

type instanceLogDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	InstanceLogID types.String `tfsdk:"instance_log_id"`
	Region        types.String `tfsdk:"region"`
	Status        types.String `tfsdk:"status"`
	NodeName      types.String `tfsdk:"node_name"`
	DownloadURL   types.String `tfsdk:"download_url"`
	CreatedAt     types.String `tfsdk:"created_at"`
	ExpiresAt     types.String `tfsdk:"expires_at"`
}

//go:embed descriptions/instance_log_data_source.md
var instanceLogDataSourceDescription string

func (d *InstanceLogDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_instance_log"
}

func (d *InstanceLogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: instanceLogDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"instance_log_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the Database Instance log. Can be a plain UUID or a regional ID (`{region}/{id}`).",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region of the Database Instance log.",
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(regional.AllRegions()),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Database Instance log, in the `{region}/{id}` format.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the log (`unknown`, `ready`, `creating`, `error`).",
			},
			"node_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the underlying node.",
			},
			"download_url": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Presigned Object Storage URL to download the log file.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation date of the log (RFC 3339 format).",
			},
			"expires_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Expiration date of the log (RFC 3339 format).",
			},
		},
	}
}

func (d *InstanceLogDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.meta = m
	d.rdbAPI = newAPI(m)
}

func (d *InstanceLogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config instanceLogDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, instanceLogID, err := resolveRegionAndID(config.InstanceLogID.ValueString(), config.Region.ValueString(), d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region and instance_log_id", err.Error())

		return
	}

	res, err := d.rdbAPI.GetInstanceLog(&rdb.GetInstanceLogRequest{
		Region:        region,
		InstanceLogID: instanceLogID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get RDB instance log",
			fmt.Sprintf("Could not retrieve instance log %s: %s", instanceLogID, err),
		)

		return
	}

	state := config
	state.InstanceLogID = types.StringValue(regional.NewIDString(region, res.ID))
	state.ID = types.StringValue(regional.NewIDString(region, res.ID))
	state.Region = types.StringValue(region.String())
	state.Status = types.StringValue(res.Status.String())
	state.NodeName = types.StringValue(res.NodeName)
	state.DownloadURL = flattenOptionalString(res.DownloadURL)
	state.CreatedAt = flattenOptionalTime(res.CreatedAt)
	state.ExpiresAt = flattenOptionalTime(res.ExpiresAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
