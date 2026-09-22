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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*VolumeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*VolumeDataSource)(nil)
)

func NewVolumeDataSource() datasource.DataSource {
	return &VolumeDataSource{}
}

type VolumeDataSource struct {
	api  *block.API
	meta *meta.Meta
}

type volumeDataSourceModel struct {
	Tags             types.List   `tfsdk:"tags"`
	ID               types.String `tfsdk:"id"`
	VolumeID         types.String `tfsdk:"volume_id"`
	Name             types.String `tfsdk:"name"`
	ProjectID        types.String `tfsdk:"project_id"`
	Zone             types.String `tfsdk:"zone"`
	SnapshotID       types.String `tfsdk:"snapshot_id"`
	InstanceVolumeID types.String `tfsdk:"instance_volume_id"`
	SRN              types.String `tfsdk:"srn"`
	Iops             types.Int64  `tfsdk:"iops"`
	SizeInGB         types.Int64  `tfsdk:"size_in_gb"`
}

func (d *VolumeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_volume"
}

func (d *VolumeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Scaleway Block Volume data source provides access to a specific Block Volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the volume, in the `{zone}/{uuid}` format.",
			},
			"volume_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the volume",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
					stringvalidator.ConflictsWith(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The volume name",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("volume_id")),
				},
			},
			"zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The zone you want to attach the resource to",
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(zonal.AllZones()),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the volume belongs to. Defaults to the provider's project ID.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"iops": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The maximum IO/s expected, must match available options",
			},
			"size_in_gb": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The volume size in GB",
			},
			"snapshot_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The snapshot to create the volume from",
			},
			"instance_volume_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The instance volume to create the block volume from",
			},
			"tags": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tags associated with the volume",
			},
			"srn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Scaleway Resource Name (SRN) of the volume",
			},
		},
	}
}

func (d *VolumeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *meta.Meta, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	d.meta = m
	d.api = block.NewAPI(d.meta.ScwClient())
}

func (d *VolumeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := meta.ExtractFrameworkZone(config.Zone, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve zone", err.Error())

		return
	}

	hasVolumeID := !config.VolumeID.IsNull() && !config.VolumeID.IsUnknown() && config.VolumeID.ValueString() != ""
	hasName := !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != ""

	var (
		volumeID     string
		volumeIDAttr types.String
	)

	switch {
	case hasVolumeID:
		rawVolumeID := config.VolumeID.ValueString()
		zonedID := scwdatasource.NewZonedID(rawVolumeID, zone)
		volumeID = zonal.ExpandID(rawVolumeID).ID
		volumeIDAttr = types.StringValue(zonedID)
	case hasName:
		volumeName := config.Name.ValueString()
		listReq := &block.ListVolumesRequest{
			Zone: zone,
			Name: &volumeName,
		}

		projectID, err := meta.ExtractFrameworkProjectID(config.ProjectID, d.meta.ScwClient())
		if err != nil {
			resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

			return
		}

		if !config.ProjectID.IsNull() {
			listReq.ProjectID = &projectID
		}

		res, listErr := d.api.ListVolumes(listReq, scw.WithContext(ctx))
		if listErr != nil {
			resp.Diagnostics.AddError("Failed to list block volumes", listErr.Error())

			return
		}

		foundVolume, findErr := scwdatasource.FindExact(
			res.Volumes,
			func(v *block.Volume) bool { return v.Name == volumeName },
			volumeName,
		)
		if findErr != nil {
			resp.Diagnostics.AddError("Failed to find block volume", findErr.Error())

			return
		}

		volumeID = foundVolume.ID
		volumeIDAttr = types.StringValue(zonal.NewIDString(zone, volumeID))
	default:
		resp.Diagnostics.AddError(
			"Missing lookup attribute",
			"Either volume_id or name must be specified.",
		)

		return
	}

	volume, err := waitForBlockVolume(ctx, d.api, zone, volumeID, defaultBlockTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read block volume", err.Error())

		return
	}

	flat := flattenVolume(ctx, d.api, volume, nil, &resp.Diagnostics)

	state := volumeDataSourceModel{
		ID:               flat.ID,
		VolumeID:         volumeIDAttr,
		Name:             flat.Name,
		ProjectID:        flat.ProjectID,
		Zone:             flat.Zone,
		Iops:             flat.Iops,
		SizeInGB:         flat.SizeInGB,
		SnapshotID:       flat.SnapshotID,
		InstanceVolumeID: flat.InstanceVolumeID,
		Tags:             flat.Tags,
		SRN:              flat.SRN,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
