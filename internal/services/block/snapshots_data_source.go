package block

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*SnapshotsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*SnapshotsDataSource)(nil)
)

func NewSnapshotsDataSource() datasource.DataSource {
	return &SnapshotsDataSource{}
}

type SnapshotsDataSource struct {
	api  *block.API
	meta *meta.Meta
}

type snapshotsDataSourceModel struct {
	Name           types.String `tfsdk:"name"`
	VolumeID       types.String `tfsdk:"volume_id"`
	Tags           types.List   `tfsdk:"tags"`
	OrderBy        types.String `tfsdk:"order_by"`
	Zone           types.String `tfsdk:"zone"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Snapshots      types.List   `tfsdk:"snapshots"`
}

func snapshotsItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":         types.StringType,
		"name":       types.StringType,
		"volume_id":  types.StringType,
		"tags":       types.ListType{ElemType: types.StringType},
		"srn":        types.StringType,
		"zone":       types.StringType,
		"project_id": types.StringType,
	}
}

func (d *SnapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_snapshots"
}

func (d *SnapshotsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Scaleway Block Storage snapshots.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter snapshots by their name",
			},
			"volume_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
				},
				MarkdownDescription: "Filter snapshots by the ID of the original volume",
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Filter snapshots by tags. Only snapshots with one or more matching tags will be returned",
			},
			"order_by": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					verify.ValidateEnumFramework[block.ListSnapshotsRequestOrderBy](),
				},
				MarkdownDescription: "Criteria to use when ordering the list of snapshots. Default value: created_at_asc",
			},
			"zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The zone in which to list snapshots. Defaults to the zone of the provider",
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(zonal.AllZones()),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter snapshots by project ID",
			},
			"organization_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter snapshots by organization ID",
			},
			"snapshots": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of snapshots",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID of the snapshot in zone/uuid format",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the snapshot",
						},
						"volume_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID of the volume from which the snapshot was created",
						},
						"tags": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "List of tags assigned to the snapshot",
						},
						"srn": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The Scaleway Resource Name (SRN) of the snapshot",
						},
						"zone": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The zone of the snapshot",
						},
						"project_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The project ID the snapshot belongs to",
						},
					},
				},
			},
		},
	}
}

func (d *SnapshotsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = block.NewAPI(m.ScwClient())
}

func (d *SnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config snapshotsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := meta.ExtractFrameworkZone(config.Zone, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve zone", err.Error())

		return
	}

	listReq := &block.ListSnapshotsRequest{
		Zone:    zone,
		OrderBy: block.ListSnapshotsRequestOrderBy(config.OrderBy.ValueString()),
	}

	if !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != "" {
		name := strings.TrimSpace(config.Name.ValueString())
		listReq.Name = &name
	}

	if !config.ProjectID.IsNull() && !config.ProjectID.IsUnknown() && config.ProjectID.ValueString() != "" {
		projectID := config.ProjectID.ValueString()
		listReq.ProjectID = &projectID
	}

	if !config.OrganizationID.IsNull() && !config.OrganizationID.IsUnknown() && config.OrganizationID.ValueString() != "" {
		organizationID := config.OrganizationID.ValueString()
		listReq.OrganizationID = &organizationID
	}

	listReq.VolumeID = scwtypes.ExpandRawID(config.VolumeID, "volume_id", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	listReq.Tags = scwtypes.ExpandStringList(ctx, config.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	listResp, err := d.api.ListSnapshots(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list block snapshots", err.Error())

		return
	}

	state := config
	state.Zone = types.StringValue(zone.String())
	state.Snapshots = flattenSnapshotsList(ctx, listResp.Snapshots, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenSnapshotsList(ctx context.Context, snapshots []*block.Snapshot, diags *diag.Diagnostics) types.List {
	itemType := types.ObjectType{AttrTypes: snapshotsItemAttrTypes()}

	if len(snapshots) == 0 {
		emptyList, d := types.ListValue(itemType, []attr.Value{})
		diags.Append(d...)

		return emptyList
	}

	items := make([]attr.Value, len(snapshots))

	for i, snapshot := range snapshots {
		tags, d := types.ListValueFrom(ctx, types.StringType, snapshot.Tags)
		diags.Append(d...)

		volumeID := types.StringNull()
		if snapshot.ParentVolume != nil {
			volumeID = types.StringValue(snapshot.ParentVolume.ID)
		}

		attrValues := map[string]attr.Value{
			"id":         types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ID)),
			"name":       types.StringValue(snapshot.Name),
			"volume_id":  volumeID,
			"tags":       tags,
			"srn":        types.StringValue(snapshot.Srn),
			"zone":       types.StringValue(string(snapshot.Zone)),
			"project_id": types.StringValue(snapshot.ProjectID),
		}

		obj, d := types.ObjectValue(itemType.AttrTypes, attrValues)
		diags.Append(d...)

		items[i] = obj
	}

	list, d := types.ListValue(itemType, items)
	diags.Append(d...)

	return list
}
