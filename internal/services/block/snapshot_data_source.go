package block

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	_ datasource.DataSource              = (*SnapshotDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*SnapshotDataSource)(nil)
)

func NewSnapshotDataSource() datasource.DataSource {
	return &SnapshotDataSource{}
}

type SnapshotDataSource struct {
	blockAPI *block.API
	meta     *meta.Meta
}

type snapshotDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	VolumeID   types.String `tfsdk:"volume_id"`
	Tags       types.List   `tfsdk:"tags"`
	SRN        types.String `tfsdk:"srn"`
	Zone       types.String `tfsdk:"zone"`
	ProjectID  types.String `tfsdk:"project_id"`
	SnapshotID types.String `tfsdk:"snapshot_id"`
}

func (d *SnapshotDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_snapshot"
}

func (d *SnapshotDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Scaleway Block Snapshot.",
		Attributes: map[string]schema.Attribute{
			"snapshot_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the snapshot",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
					stringvalidator.ConflictsWith(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The snapshot name",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("snapshot_id")),
				},
			},
			"volume_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the volume from which creates a snapshot",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The tags associated with the snapshot",
			},
			"srn": schema.StringAttribute{
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the snapshot",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The zone you want to attach the resource to",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The project ID the snapshot belongs to.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the snapshot, in the `{zone}/{id}` format.",
			},
		},
	}
}

func (d *SnapshotDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.blockAPI = block.NewAPI(d.meta.ScwClient())
}

func (d *SnapshotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config snapshotDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := meta.ExtractFrameworkZone(config.Zone, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve zone", err.Error())

		return
	}

	var snapshotID string

	if config.SnapshotID.IsNull() || config.SnapshotID.ValueString() == "" {
		name := config.Name.ValueString()

		listReq := &block.ListSnapshotsRequest{
			Zone:     zone,
			Name:     scwtypes.ExpandStringPtr(name),
			VolumeID: scwtypes.ExpandStringPtr(config.VolumeID.ValueString()),
		}

		if !config.ProjectID.IsNull() && config.ProjectID.ValueString() != "" {
			listReq.ProjectID = scwtypes.ExpandStringPtr(config.ProjectID.ValueString())
		}

		res, err := d.blockAPI.ListSnapshots(listReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to list block snapshots", err.Error())

			return
		}

		for _, snapshot := range res.Snapshots {
			if snapshot.Name == name {
				if snapshotID != "" {
					resp.Diagnostics.AddError("Duplicate snapshot found",
						"More than 1 snapshot found with the same name "+name)

					return
				}

				snapshotID = snapshot.ID
			}
		}

		if snapshotID == "" {
			resp.Diagnostics.AddError("Snapshot not found",
				"No snapshot found with the name "+name)

			return
		}
	} else {
		snapshotID = zonal.ExpandID(config.SnapshotID.ValueString()).ID
	}

	snapshot, err := d.blockAPI.GetSnapshot(&block.GetSnapshotRequest{
		Zone:       zone,
		SnapshotID: snapshotID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get block snapshot", err.Error())

		return
	}

	state := flattenSnapshotDataSource(ctx, snapshot, config, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenSnapshotDataSource(ctx context.Context, snapshot *block.Snapshot, config snapshotDataSourceModel, diags *diag.Diagnostics) snapshotDataSourceModel {
	model := snapshotDataSourceModel{
		ID:         types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ID)),
		SnapshotID: types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ID)),
		Name:       types.StringValue(snapshot.Name),
		ProjectID:  types.StringValue(snapshot.ProjectID),
		Zone:       types.StringValue(snapshot.Zone.String()),
		SRN:        types.StringValue(snapshot.Srn),
	}

	tagsList, d := scwtypes.FlattenStringList(ctx, "tags", snapshot.Tags, config)
	diags.Append(d...)

	model.Tags = tagsList

	if snapshot.ParentVolume != nil {
		model.VolumeID = types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ParentVolume.ID))
	} else {
		model.VolumeID = types.StringNull()
	}

	return model
}
