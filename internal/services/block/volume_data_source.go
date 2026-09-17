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
	_ datasource.DataSource              = (*VolumeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*VolumeDataSource)(nil)
)

func NewVolumeDataSource() datasource.DataSource {
	return &VolumeDataSource{}
}

type VolumeDataSource struct {
	blockAPI *block.API
	meta     *meta.Meta
}

type volumeDataSourceModel struct {
	Tags             types.List   `tfsdk:"tags"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	SnapshotID       types.String `tfsdk:"snapshot_id"`
	InstanceVolumeID types.String `tfsdk:"instance_volume_id"`
	SRN              types.String `tfsdk:"srn"`
	Zone             types.String `tfsdk:"zone"`
	ProjectID        types.String `tfsdk:"project_id"`
	VolumeID         types.String `tfsdk:"volume_id"`
	Iops             types.Int64  `tfsdk:"iops"`
	SizeInGB         types.Int64  `tfsdk:"size_in_gb"`
}

func (d *VolumeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_volume"
}

func (d *VolumeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Scaleway Block Volume.",
		Attributes: map[string]schema.Attribute{
			"volume_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the volume",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
					stringvalidator.ConflictsWith(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The volume name",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("volume_id")),
				},
			},
			"iops": schema.Int64Attribute{
				Computed:    true,
				Description: "The maximum IO/s expected, must match available options",
			},
			"size_in_gb": schema.Int64Attribute{
				Computed:    true,
				Description: "The volume size in GB",
			},
			"snapshot_id": schema.StringAttribute{
				Computed:    true,
				Description: "The snapshot to create the volume from",
			},
			"instance_volume_id": schema.StringAttribute{
				Computed:    true,
				Description: "The instance volume to create the block volume from",
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The tags associated with the volume",
			},
			"srn": schema.StringAttribute{
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the volume",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The zone you want to attach the resource to",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The project ID the volume belongs to.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the volume, in the `{zone}/{id}` format.",
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
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.meta = m
	d.blockAPI = block.NewAPI(d.meta.ScwClient())
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

	var volumeID string

	if config.VolumeID.IsNull() || config.VolumeID.ValueString() == "" {
		name := config.Name.ValueString()

		listReq := &block.ListVolumesRequest{
			Zone: zone,
			Name: scwtypes.ExpandStringPtr(name),
		}

		if !config.ProjectID.IsNull() && config.ProjectID.ValueString() != "" {
			listReq.ProjectID = scwtypes.ExpandStringPtr(config.ProjectID.ValueString())
		}

		res, err := d.blockAPI.ListVolumes(listReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to list block volumes", err.Error())

			return
		}

		for _, volume := range res.Volumes {
			if volume.Name == name {
				if volumeID != "" {
					resp.Diagnostics.AddError("Duplicate volume found",
						"More than 1 volume found with the same name "+name)

					return
				}

				volumeID = volume.ID
			}
		}

		if volumeID == "" {
			resp.Diagnostics.AddError("Volume not found",
				"No volume found with the name "+name)

			return
		}
	} else {
		volumeID = zonal.ExpandID(config.VolumeID.ValueString()).ID
	}

	volume, err := d.blockAPI.GetVolume(&block.GetVolumeRequest{
		Zone:     zone,
		VolumeID: volumeID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get block volume", err.Error())

		return
	}

	state := flattenVolumeDataSource(ctx, volume, config, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenVolumeDataSource(ctx context.Context, volume *block.Volume, config volumeDataSourceModel, diags *diag.Diagnostics) volumeDataSourceModel {
	model := volumeDataSourceModel{
		ID:        types.StringValue(zonal.NewIDString(volume.Zone, volume.ID)),
		VolumeID:  types.StringValue(zonal.NewIDString(volume.Zone, volume.ID)),
		Name:      types.StringValue(volume.Name),
		SizeInGB:  types.Int64Value(int64(volume.Size / scw.GB)),
		ProjectID: types.StringValue(volume.ProjectID),
		Zone:      types.StringValue(volume.Zone.String()),
		SRN:       types.StringValue(volume.Srn),
	}

	tagsList, d := scwtypes.FlattenStringList(ctx, "tags", volume.Tags, config)
	diags.Append(d...)

	model.Tags = tagsList

	if volume.Specs != nil && volume.Specs.PerfIops != nil {
		model.Iops = types.Int64Value(int64(*volume.Specs.PerfIops))
	} else {
		model.Iops = types.Int64Value(0)
	}

	if volume.ParentSnapshotID != nil {
		model.SnapshotID = types.StringValue(zonal.NewIDString(volume.Zone, *volume.ParentSnapshotID))
	} else {
		model.SnapshotID = types.StringNull()
	}

	return model
}
