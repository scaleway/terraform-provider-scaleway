package block

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*VolumeResource)(nil)
	_ resource.ResourceWithConfigure   = (*VolumeResource)(nil)
	_ resource.ResourceWithImportState = (*VolumeResource)(nil)
	_ resource.ResourceWithIdentity    = (*VolumeResource)(nil)
)

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

type VolumeResource struct {
	api  *block.API
	meta *meta.Meta
}

type volumeResourceModel struct {
	ID               types.String `tfsdk:"id"`
	InstanceVolumeID types.String `tfsdk:"instance_volume_id"`
	Iops             types.Int64  `tfsdk:"iops"`
	Name             types.String `tfsdk:"name"`
	ProjectID        types.String `tfsdk:"project_id"`
	SRN              types.String `tfsdk:"srn"`
	SizeInGB         types.Int64  `tfsdk:"size_in_gb"`
	SnapshotID       types.String `tfsdk:"snapshot_id"`
	Tags             types.List   `tfsdk:"tags"`
	Zone             types.String `tfsdk:"zone"`
}

func (r *VolumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_volume"
}

func (r *VolumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway Block Volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the volume, in the `{zone}/{uuid} format.`",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The volume name",
			},
			"iops": schema.Int64Attribute{
				Required:    true,
				Description: "The maximum IO/s expected, must match available options",
			},
			"size_in_gb": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The volume size in GB",
			},
			"snapshot_id": schema.StringAttribute{
				Optional:    true,
				Description: "The snapshot to create the volume from",
				PlanModifiers: []planmodifier.String{
					zonal.LocalityPlanModifier(),
				},
			},
			"instance_volume_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The instance volume to create the block volume from",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("snapshot_id")),
				},
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "The tags associated with the volume",
			},
			"srn": schema.StringAttribute{
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the volume",
			},
			"zone": zonal.SchemaAttributeComputed(
				"The zone you want to attach the resource to",
			),
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the volume belongs to. Defaults to the provider's project ID.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VolumeResource) IdentitySchema(
	_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse,
) {
	resp.IdentitySchema = framework.DefaultZonal()
}

func (r *VolumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
	r.api = block.NewAPI(r.meta.ScwClient())
}

func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data volumeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := meta.ExtractFrameworkZone(data.Zone, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve zone", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	var volume *block.Volume

	if !data.InstanceVolumeID.IsNull() && data.InstanceVolumeID.ValueString() != "" {
		// Volume is from an instance volume, import
		instanceAPI := instancehelpers.NewBlockAndInstanceAPI(r.meta.ScwClient())

		volume, err = migrateInstanceToBlockVolume(
			ctx, instanceAPI, zone, locality.ExpandID(data.InstanceVolumeID.ValueString()), defaultBlockTimeout,
		)
	} else {
		// Volume is new, create
		createReq := &block.CreateVolumeRequest{
			Zone:      zone,
			Name:      scwtypes.ExpandOrGenerateString(data.Name.ValueString(), "volume"),
			ProjectID: projectID,
			PerfIops:  new(uint32(data.Iops.ValueInt64())),
		}

		createReq.Tags = scwtypes.ExpandUpdatedStringList(ctx, data.Tags, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		if !data.SnapshotID.IsNull() && data.SnapshotID.ValueString() != "" {
			createReq.FromSnapshot = &block.CreateVolumeRequestFromSnapshot{
				SnapshotID: locality.ExpandID(data.SnapshotID.ValueString()),
			}
		}

		if !data.SizeInGB.IsNull() && !data.SizeInGB.IsUnknown() {
			volumeSizeInBytes := scw.Size(data.SizeInGB.ValueInt64()) * scw.GB
			if createReq.FromSnapshot != nil {
				createReq.FromSnapshot.Size = &volumeSizeInBytes
			} else {
				createReq.FromEmpty = &block.CreateVolumeRequestFromEmpty{
					Size: volumeSizeInBytes,
				}
			}
		}

		volume, err = r.api.CreateVolume(createReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to create block volume", err.Error())

			return
		}
	}

	volume, err = waitForBlockVolume(ctx, r.api, zone, volume.ID, defaultBlockTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for block volume during Create", err.Error())

		return
	}

	state := flattenVolume(ctx, volume, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(
		ctx, framework.SetZonalIdentity(volume.Zone, volume.ID),
	)...)
}

func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}

func flattenVolume(ctx context.Context, volume *block.Volume, reference any, diags *diag.Diagnostics) volumeResourceModel {
	model := volumeResourceModel{
		ID:        types.StringValue(zonal.NewIDString(volume.Zone, volume.ID)),
		Name:      types.StringValue(volume.Name),
		SizeInGB:  types.Int64Value(int64(volume.Size / scw.GB)),
		ProjectID: types.StringValue(volume.ProjectID),
		Zone:      types.StringValue(volume.Zone.String()),
		SRN:       types.StringValue(volume.Srn),
	}

	tagsList, d := scwtypes.FlattenStringList(ctx, "tags", volume.Tags, reference)
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
