package rdb

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
	rdb "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*InstanceLogsDetailsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*InstanceLogsDetailsDataSource)(nil)
)

func NewInstanceLogsDetailsDataSource() datasource.DataSource {
	return &InstanceLogsDetailsDataSource{}
}

type InstanceLogsDetailsDataSource struct {
	rdbAPI *rdb.API
	meta   *meta.Meta
}

type instanceLogsDetailsDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	InstanceID types.String `tfsdk:"instance_id"`
	Region     types.String `tfsdk:"region"`
	Details    types.List   `tfsdk:"details"`
}

func instanceLogDetailAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"log_name":      types.StringType,
		"size_in_bytes": types.Int64Type,
	}
}

//go:embed descriptions/instance_logs_details_data_source.md
var instanceLogsDetailsDataSourceDescription string

func (d *InstanceLogsDetailsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_instance_logs_details"
}

func (d *InstanceLogsDetailsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: instanceLogsDetailsDataSourceDescription,
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
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this data source, in the `{region}/{instance_id}` format.",
			},
			"details": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Remote Database Instance logs details.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"log_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the remote log.",
						},
						"size_in_bytes": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Size of the remote log in bytes.",
						},
					},
				},
			},
		},
	}
}

func (d *InstanceLogsDetailsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InstanceLogsDetailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config instanceLogsDetailsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, instanceID, err := resolveRegionAndID(config.InstanceID.ValueString(), config.Region.ValueString(), d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region and instance_id", err.Error())

		return
	}

	res, err := d.rdbAPI.ListInstanceLogsDetails(&rdb.ListInstanceLogsDetailsRequest{
		Region:     region,
		InstanceID: instanceID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list RDB instance logs details",
			fmt.Sprintf("Could not list logs details for instance %s: %s", instanceID, err),
		)

		return
	}

	state := config
	state.InstanceID = types.StringValue(regional.NewIDString(region, instanceID))
	state.Region = types.StringValue(region.String())
	state.ID = types.StringValue(regional.NewIDString(region, instanceID))
	state.Details = flattenInstanceLogsDetails(res.Details, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenInstanceLogsDetails(details []*rdb.ListInstanceLogsDetailsResponseInstanceLogDetail, diags *diag.Diagnostics) types.List {
	itemType := types.ObjectType{AttrTypes: instanceLogDetailAttrTypes()}

	if len(details) == 0 {
		emptyList, d := types.ListValue(itemType, []attr.Value{})
		diags.Append(d...)

		return emptyList
	}

	items := make([]attr.Value, 0, len(details))

	for _, detail := range details {
		if detail == nil {
			continue
		}

		obj, d := types.ObjectValue(instanceLogDetailAttrTypes(), map[string]attr.Value{
			"log_name":      types.StringValue(detail.LogName),
			"size_in_bytes": types.Int64Value(int64(detail.Size)),
		})
		diags.Append(d...)

		items = append(items, obj)
	}

	list, d := types.ListValue(itemType, items)
	diags.Append(d...)

	return list
}
