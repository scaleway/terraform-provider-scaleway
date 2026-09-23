package block

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	scwdatasource "github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	providertypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
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
	api  *block.API
	meta *meta.Meta
}

type snapshotDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	SnapshotID types.String `tfsdk:"snapshot_id"`
	Name       types.String `tfsdk:"name"`
	VolumeID   types.String `tfsdk:"volume_id"`
	Tags       types.List   `tfsdk:"tags"`
	Srn        types.String `tfsdk:"srn"`
	Zone       types.String `tfsdk:"zone"`
	ProjectID  types.String `tfsdk:"project_id"`
}

func (d *SnapshotDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_snapshot"
}

func (d *SnapshotDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway Block Snapshot",
		Attributes: map[string]schema.Attribute{
			"snapshot_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the snapshot",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
					stringvalidator.ConflictsWith(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The snapshot name",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("snapshot_id")),
				},
			},
			"volume_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID of the volume from which creates a snapshot",
			},
			"tags": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tags associated with the snapshot",
			},
			"srn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Scaleway Resource Name (SRN) of the snapshot",
			},
			"zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The zone of the snapshot",
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(zonal.AllZones()),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the snapshot belongs to",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the snapshot, in the `{zone}/{id}` format.",
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
	d.api = block.NewAPI(d.meta.ScwClient())
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

	hasSnapshotID := !config.SnapshotID.IsNull() && !config.SnapshotID.IsUnknown() && config.SnapshotID.ValueString() != ""
	hasName := !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != ""

	switch {
	case hasSnapshotID:
		snapshotID = locality.ExpandID(config.SnapshotID.ValueString())
	case hasName:
		listReq := &block.ListSnapshotsRequest{
			Zone: zone,
			Name: providertypes.ExpandStringPtr(config.Name.ValueString()),
		}

		if !config.ProjectID.IsNull() && !config.ProjectID.IsUnknown() && config.ProjectID.ValueString() != "" {
			projectID := config.ProjectID.ValueString()
			listReq.ProjectID = &projectID
		}

		if !config.VolumeID.IsNull() && !config.VolumeID.IsUnknown() && config.VolumeID.ValueString() != "" {
			volumeID := locality.ExpandID(config.VolumeID.ValueString())
			listReq.VolumeID = &volumeID
		}

		res, err := d.api.ListSnapshots(listReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to list block snapshots", err.Error())

			return
		}

		foundSnapshot, findErr := scwdatasource.FindExact(
			res.Snapshots,
			func(s *block.Snapshot) bool { return s.Name == config.Name.ValueString() },
			config.Name.ValueString(),
		)
		if findErr != nil {
			resp.Diagnostics.AddError("Failed to find block snapshot", findErr.Error())

			return
		}

		snapshotID = foundSnapshot.ID
	default:
		resp.Diagnostics.AddError(
			"Missing lookup attribute",
			"Either snapshot_id or name must be specified.",
		)

		return
	}

	snapshot, err := d.api.GetSnapshot(&block.GetSnapshotRequest{
		Zone:       zone,
		SnapshotID: snapshotID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read block snapshot", err.Error())

		return
	}

	flat := flattenBlockSnapshot(ctx, snapshot, &config, &resp.Diagnostics)

	state := snapshotDataSourceModel{
		ID:         flat.ID,
		Name:       flat.Name,
		Tags:       flat.Tags,
		Srn:        flat.Srn,
		Zone:       flat.Zone,
		ProjectID:  flat.ProjectID,
		SnapshotID: types.StringValue(zonal.NewIDString(snapshot.Zone, snapshot.ID)),
	}

	if snapshot.ParentVolume != nil {
		state.VolumeID = types.StringValue(snapshot.ParentVolume.ID)
	} else {
		state.VolumeID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
