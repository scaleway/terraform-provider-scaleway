package messageq

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

//go:embed descriptions/node_type_data_source.md
var nodeTypeDataSourceDescription string

var (
	_ datasource.DataSource              = (*NodeTypeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*NodeTypeDataSource)(nil)
)

func NewNodeTypeDataSource() datasource.DataSource {
	return &NodeTypeDataSource{}
}

type NodeTypeDataSource struct {
	api  *messageqapi.API
	meta *meta.Meta
}

type nodeTypeDataSourceModel struct {
	AvailableVolumeTypes types.List   `tfsdk:"available_volume_types"`
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Region               types.String `tfsdk:"region"`
	Description          types.String `tfsdk:"description"`
	StockStatus          types.String `tfsdk:"stock_status"`
	InstanceRange        types.String `tfsdk:"instance_range"`
	Vcpus                types.Int64  `tfsdk:"vcpus"`
	MemorySizeInGB       types.Int64  `tfsdk:"memory_size_in_gb"`
	Disabled             types.Bool   `tfsdk:"disabled"`
	Beta                 types.Bool   `tfsdk:"beta"`
}

func availableVolumeTypeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":             types.StringType,
		"description":      types.StringType,
		"min_size_in_gb":   types.Int64Type,
		"max_size_in_gb":   types.Int64Type,
		"chunk_size_in_gb": types.Int64Type,
	}
}

func (d *NodeTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_messageq_node_type"
}

func (d *NodeTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: nodeTypeDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The node type name",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region the node type is available in.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the node type, in the `{region}/{name}` format.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the node type",
			},
			"vcpus": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of vCPUs available",
			},
			"memory_size_in_gb": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Amount of memory available in GB",
			},
			"stock_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Stock status of the node type",
			},
			"disabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the node type is disabled",
			},
			"beta": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the node type is in beta",
			},
			"instance_range": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Instance range associated with the node type offer",
			},
			"available_volume_types": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Available storage options for the node type",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Volume type",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Volume type description",
						},
						"min_size_in_gb": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Minimum volume size in GB",
						},
						"max_size_in_gb": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Maximum volume size in GB",
						},
						"chunk_size_in_gb": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Volume size increment in GB",
						},
					},
				},
			},
		},
	}
}

func (d *NodeTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = messageqapi.NewAPI(d.meta.ScwClient())
}

func (d *NodeTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config nodeTypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	name := config.Name.ValueString()

	res, err := d.api.ListNodeTypes(&messageqapi.ListNodeTypesRequest{
		Region: region,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list MessageQ node types", err.Error())

		return
	}

	var found *messageqapi.NodeType

	for _, nodeType := range res.NodeTypes {
		if nodeType.Name == name {
			found = nodeType

			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("MessageQ node type not found", fmt.Sprintf("messageq node type %q not found", name))

		return
	}

	state := nodeTypeDataSourceModel{
		ID:             types.StringValue(regional.NewIDString(region, found.Name)),
		Region:         types.StringValue(region.String()),
		Name:           types.StringValue(found.Name),
		Description:    types.StringValue(found.Description),
		Vcpus:          types.Int64Value(int64(found.Vcpus)),
		MemorySizeInGB: types.Int64Value(int64(BytesToGB(found.MemoryBytes))),
		StockStatus:    types.StringValue(string(found.StockStatus)),
		Disabled:       types.BoolValue(found.Disabled),
		Beta:           types.BoolValue(found.Beta),
		InstanceRange:  types.StringValue(found.InstanceRange),
	}

	state.AvailableVolumeTypes = flattenAvailableVolumeTypes(found.AvailableVolumeTypes, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenAvailableVolumeTypes(volumeTypes []*messageqapi.NodeTypeVolumeType, diags *diag.Diagnostics) types.List {
	elemType := types.ObjectType{AttrTypes: availableVolumeTypeAttrTypes()}

	if len(volumeTypes) == 0 {
		return types.ListNull(elemType)
	}

	values := make([]attr.Value, 0, len(volumeTypes))

	for _, volumeType := range volumeTypes {
		obj, d := types.ObjectValue(availableVolumeTypeAttrTypes(), map[string]attr.Value{
			"type":             types.StringValue(string(volumeType.Type)),
			"description":      types.StringValue(volumeType.Description),
			"min_size_in_gb":   types.Int64Value(int64(BytesToGB(volumeType.MinSizeBytes))),
			"max_size_in_gb":   types.Int64Value(int64(BytesToGB(volumeType.MaxSizeBytes))),
			"chunk_size_in_gb": types.Int64Value(int64(BytesToGB(volumeType.ChunkSizeBytes))),
		})
		diags.Append(d...)

		values = append(values, obj)
	}

	listVal, d := types.ListValue(elemType, values)
	diags.Append(d...)

	return listVal
}
