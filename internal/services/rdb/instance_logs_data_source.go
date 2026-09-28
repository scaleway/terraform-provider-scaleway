package rdb

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	rdb "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*InstanceLogsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*InstanceLogsDataSource)(nil)
)

func NewInstanceLogsDataSource() datasource.DataSource {
	return &InstanceLogsDataSource{}
}

type InstanceLogsDataSource struct {
	rdbAPI *rdb.API
	meta   *meta.Meta
}

type instanceLogsDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	InstanceID   types.String `tfsdk:"instance_id"`
	Region       types.String `tfsdk:"region"`
	OrderBy      types.String `tfsdk:"order_by"`
	InstanceLogs types.List   `tfsdk:"instance_logs"`
}

func instanceLogAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":           types.StringType,
		"status":       types.StringType,
		"node_name":    types.StringType,
		"download_url": types.StringType,
		"created_at":   types.StringType,
		"expires_at":   types.StringType,
		"region":       types.StringType,
	}
}

//go:embed descriptions/instance_logs_data_source.md
var instanceLogsDataSourceDescription string

func (d *InstanceLogsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_instance_logs"
}

func (d *InstanceLogsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: instanceLogsDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"instance_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the Database Instance. Can be a plain UUID or a regional ID (`{region}/{id}`).",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region of the Database Instance.",
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(regional.AllRegions()),
				},
			},
			"order_by": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Criteria to use when ordering Database Instance logs listing. Possible values are `created_at_asc` and `created_at_desc`.",
				Validators: []validator.String{
					verify.FrameworkValidateEnum[rdb.ListInstanceLogsRequestOrderBy](),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this data source, in the `{region}/{instance_id}` format.",
			},
			"instance_logs": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available logs for the Database Instance.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: instanceLogSchemaAttributes(),
				},
			},
		},
	}
}

func instanceLogSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "UUID of the Database Instance log, in the `{region}/{id}` format.",
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
		"region": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Region the Database Instance is in.",
		},
	}
}

func (d *InstanceLogsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InstanceLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config instanceLogsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, instanceID, err := resolveRegionAndID(config.InstanceID.ValueString(), config.Region.ValueString(), d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region and instance_id", err.Error())

		return
	}

	listReq := &rdb.ListInstanceLogsRequest{
		Region:     region,
		InstanceID: instanceID,
	}

	if !config.OrderBy.IsNull() && !config.OrderBy.IsUnknown() && config.OrderBy.ValueString() != "" {
		listReq.OrderBy = rdb.ListInstanceLogsRequestOrderBy(config.OrderBy.ValueString())
	}

	res, err := d.rdbAPI.ListInstanceLogs(listReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list RDB instance logs",
			fmt.Sprintf("Could not list logs for instance %s: %s", instanceID, err),
		)

		return
	}

	state := config
	state.InstanceID = types.StringValue(regional.NewIDString(region, instanceID))
	state.Region = types.StringValue(region.String())
	state.ID = types.StringValue(regional.NewIDString(region, instanceID))
	state.InstanceLogs = flattenInstanceLogs(res.InstanceLogs, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenInstanceLogs(logs []*rdb.InstanceLog, diags *diag.Diagnostics) types.List {
	itemType := types.ObjectType{AttrTypes: instanceLogAttrTypes()}

	if len(logs) == 0 {
		emptyList, d := types.ListValue(itemType, []attr.Value{})
		diags.Append(d...)

		return emptyList
	}

	items := make([]attr.Value, 0, len(logs))

	for _, log := range logs {
		if log == nil {
			continue
		}

		obj, d := flattenInstanceLogObject(log)
		diags.Append(d...)

		items = append(items, obj)
	}

	list, d := types.ListValue(itemType, items)
	diags.Append(d...)

	return list
}

func flattenInstanceLogObject(log *rdb.InstanceLog) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	logID := log.ID
	if log.Region != "" {
		logID = regional.NewIDString(log.Region, log.ID)
	}

	attrValues := map[string]attr.Value{
		"id":           types.StringValue(logID),
		"status":       types.StringValue(log.Status.String()),
		"node_name":    types.StringValue(log.NodeName),
		"download_url": flattenOptionalString(log.DownloadURL),
		"created_at":   flattenOptionalTime(log.CreatedAt),
		"expires_at":   flattenOptionalTime(log.ExpiresAt),
		"region":       types.StringValue(log.Region.String()),
	}

	obj, d := types.ObjectValue(instanceLogAttrTypes(), attrValues)
	diags.Append(d...)

	return obj, diags
}

func flattenOptionalString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}

	return types.StringValue(*s)
}

func flattenOptionalTime(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}

	return types.StringValue(t.Format(time.RFC3339))
}
