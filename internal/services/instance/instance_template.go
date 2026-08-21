package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	instanceV2 "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*InstanceTemplateResource)(nil)
	_ resource.ResourceWithConfigure   = (*InstanceTemplateResource)(nil)
	_ resource.ResourceWithImportState = (*InstanceTemplateResource)(nil)
)

func NewInstanceTemplateResource() resource.Resource {
	return &InstanceTemplateResource{}
}

type InstanceTemplateResource struct {
	api  *instanceV2.API
	meta *meta.Meta
}

type instanceTemplateResourceModel struct {
	Volumes            types.List   `tfsdk:"volumes"`
	FilesystemIDs      types.Set    `tfsdk:"filesystem_ids"`
	PrivateNetworks    types.Set    `tfsdk:"private_networks"`
	Tags               types.List   `tfsdk:"tags"`
	ServerTags         types.List   `tfsdk:"server_tags"`
	CreatedAt          types.String `tfsdk:"created_at"`
	SecurityGroupID    types.String `tfsdk:"security_group_id"`
	PlacementGroupID   types.String `tfsdk:"placement_group_id"`
	ServerType         types.String `tfsdk:"server_type"`
	Name               types.String `tfsdk:"name"`
	ProjectID          types.String `tfsdk:"project_id"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
	WindowsRdpSSHKeyID types.String `tfsdk:"windows_rdp_ssh_key_id"`
	ID                 types.String `tfsdk:"id"`
	Zone               types.String `tfsdk:"zone"`
	PublicIPV4Count    types.Int32  `tfsdk:"public_ipv4_count"`
	PublicIPV6Count    types.Int32  `tfsdk:"public_ipv6_count"`
}

func (r *InstanceTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_template"
}

func (r *InstanceTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway Instance Template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Instance Template, in the `{zone}/{id}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the Instance Template. If not provided, a random name will be generated.",
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tags associated with the Instance Template.",
			},
			"server_tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tags that will be assigned to the servers created using the Instance Template.",
			},
			"server_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The commercial type of the server defined by the Instance Template.",
			},
			"security_group_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the security group to attach to the servers created using the Instance Template.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
				},
			},
			"placement_group_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the placement group to attach to the servers created using the Instance Template.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
				},
			},
			"public_ipv4_count": schema.Int32Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(0),
				MarkdownDescription: "The number of public IPv4 to attach to the servers created using the Instance Template.",
			},
			"public_ipv6_count": schema.Int32Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(0),
				MarkdownDescription: "The number of public IPv6 to attach to the servers created using the Instance Template.",
			},
			"volumes": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "The specs of the volumes of the servers created using the Instance Template.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"volume_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The type of volume.",
							Validators: []validator.String{
								verify.ValidateEnumFramework[instanceV2.ServerVolumeVolumeType](),
							},
						},
						"size_in_gb": schema.Int64Attribute{
							Required:            true,
							MarkdownDescription: "The size of the volume in gigabytes.",
						},
						"name": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "The name of volume. If not provided, a random name will be generated.",
						},
						"tags": schema.ListAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "The tags associated with the volume.",
						},
						"base_snapshot_id": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The ID of the base snapshot for the volume.",
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("image_label")),
								verify.IsStringUUIDOrUUIDWithZone(),
							},
						},
						"image_label": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The label of the image used as base for the volume.",
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("base_snapshot_id")),
							},
						},
						"perf_iops": schema.Int32Attribute{
							Optional:            true,
							MarkdownDescription: "The performance IOPS of the volume.",
						},
					},
				},
			},
			"private_networks": schema.SetAttribute{
				Optional:            true,
				MarkdownDescription: "The IDs of the private networks to attach to the servers created using the Instance Template.",
				ElementType:         types.StringType,
				Validators: []validator.Set{
					verify.SetElemIsStringUUIDOrUUIDWithRegion(),
				},
			},
			"windows_rdp_ssh_key_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the IAM SSH key used to encrypt the initial admin password on a Windows server. This will be repeated on all servers created using the Instance Template.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"filesystem_ids": schema.SetAttribute{
				Optional:            true,
				MarkdownDescription: "The IDs of the filesystems to attach to the servers created using the Instance Template.",
				ElementType:         types.StringType,
				Validators: []validator.Set{
					verify.SetElemIsStringUUIDOrUUIDWithRegion(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creation timestamp of the Instance Template.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The last update timestamp of the Instance Template.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the Instance Template belongs to. Defaults to the provider's project ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The zone the Instance Template is in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *InstanceTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = instanceV2.NewAPI(r.meta.ScwClient())
}

func (r *InstanceTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data instanceTemplateResourceModel
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

	createReq := &instanceV2.CreateTemplateRequest{
		Zone:            zone,
		ProjectID:       projectID,
		Name:            scwtypes.ExpandOrGenerateString(data.Name.ValueString(), "tf-tmpl"),
		ServerType:      data.ServerType.ValueString(),
		PublicIPV4Count: uint32(data.PublicIPV4Count.ValueInt32()),
		PublicIPV6Count: uint32(data.PublicIPV6Count.ValueInt32()),
	}

	createReq.Tags = scwtypes.ExpandStringList(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.ServerTags = scwtypes.ExpandStringList(ctx, data.ServerTags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.SecurityGroupID = scwtypes.ExpandRawID(data.SecurityGroupID, "security_group_id", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.PlacementGroupID = scwtypes.ExpandRawID(data.PlacementGroupID, "placement_group_id", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.WindowsRdpSSHKeyID = scwtypes.ExpandRawID(data.WindowsRdpSSHKeyID, "windows_rdp_ssh_key_id", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.Volumes = expandVolumes(ctx, data.Volumes, &resp.Diagnostics).ToCreateRequest()
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.PrivateNetworks = expandPrivateNetworks(ctx, data.PrivateNetworks, &resp.Diagnostics).ToCreateRequest()
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.FilesystemIDs = scwtypes.ExpandRawIDSet(ctx, data.FilesystemIDs, "filesystem_ids", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.api.CreateTemplate(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Instance Template", err.Error())

		return
	}

	state := flattenInstanceTemplate(ctx, tmpl, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenInstanceTemplate(ctx context.Context, tmpl *instanceV2.Template, reference any, diags *diag.Diagnostics) any {
	model := instanceTemplateResourceModel{
		ProjectID:       types.StringValue(tmpl.ProjectID),
		ID:              types.StringValue(zonal.NewIDString(tmpl.Zone, tmpl.ID)),
		Name:            types.StringValue(tmpl.Name),
		ServerType:      types.StringValue(tmpl.ServerType),
		PublicIPV4Count: types.Int32Value(int32(tmpl.PublicIPV4Count)),
		PublicIPV6Count: types.Int32Value(int32(tmpl.PublicIPV6Count)),
		Zone:            types.StringValue(tmpl.Zone.String()),
	}

	tagList, d := scwtypes.FlattenStringList(ctx, tmpl.Tags)
	diags.Append(d...)

	model.Tags = tagList

	serverTagList, d := scwtypes.FlattenStringList(ctx, tmpl.ServerTags)
	diags.Append(d...)

	model.ServerTags = serverTagList

	if tmpl.SecurityGroupID != nil {
		if scwtypes.IDUsesZonedFormat(ctx, reference, path.Root("security_group_id"), diags) {
			model.SecurityGroupID = types.StringValue(zonal.NewIDString(tmpl.Zone, *tmpl.SecurityGroupID))
		} else {
			model.SecurityGroupID = types.StringValue(*tmpl.SecurityGroupID)
		}
	} else {
		model.SecurityGroupID = types.StringNull()
	}

	if tmpl.PlacementGroupID != nil {
		if scwtypes.IDUsesZonedFormat(ctx, reference, path.Root("placement_group_id"), diags) {
			model.PlacementGroupID = types.StringValue(zonal.NewIDString(tmpl.Zone, *tmpl.PlacementGroupID))
		} else {
			model.PlacementGroupID = types.StringValue(*tmpl.PlacementGroupID)
		}
	} else {
		model.PlacementGroupID = types.StringNull()
	}

	volumesList, d := flattenVolumes(ctx, tmpl.Volumes, reference, tmpl.Zone)
	diags.Append(d...)

	model.Volumes = volumesList

	privateNetworkIDsList, d := flattenPrivateNetworks(ctx, tmpl.PrivateNetworks, tmpl.Zone, reference)
	diags.Append(d...)

	model.PrivateNetworks = privateNetworkIDsList

	if tmpl.WindowsRdpSSHKeyID != nil {
		model.WindowsRdpSSHKeyID = types.StringValue(*tmpl.WindowsRdpSSHKeyID)
	} else {
		model.WindowsRdpSSHKeyID = types.StringNull()
	}

	if tmpl.CreatedAt != nil {
		model.CreatedAt = types.StringValue(tmpl.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		model.CreatedAt = types.StringNull()
	}

	if tmpl.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(tmpl.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	regionStr := ""

	region, err := tmpl.Zone.Region()
	if err != nil {
		diags.Append(diag.NewAttributeWarningDiagnostic(path.Root("filesystem_ids"), "localized IDs may be malformed", fmt.Sprintf("failed to infer region from template zone %q: %s", tmpl.Zone, err.Error())))
	} else {
		regionStr = region.String()
	}

	filesystemIDsList, d := scwtypes.FlattenIDSet(ctx, "filesystem_ids", tmpl.FilesystemIDs, regionStr, reference)
	diags.Append(d...)

	model.FilesystemIDs = filesystemIDsList

	return model
}

func (r *InstanceTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Instance Template ID", err.Error())

		return
	}

	tmpl, err := r.api.GetTemplate(&instanceV2.GetTemplateRequest{
		Zone:       zone,
		TemplateID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read Instance Template", err.Error())

		return
	}

	newState := flattenInstanceTemplate(ctx, tmpl, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, new(newState))...)
}

func (r *InstanceTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  instanceTemplateResourceModel
		state instanceTemplateResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Instance Template ID", err.Error())

		return
	}

	updateReq := &instanceV2.UpdateTemplateRequest{
		Zone:       zone,
		TemplateID: id,
	}
	hasChanges := false

	if !plan.Name.Equal(state.Name) {
		updateReq.Name = new(plan.Name.ValueString())
		hasChanges = true
	}

	if !plan.Tags.Equal(state.Tags) {
		updateReq.Tags = new(scwtypes.ExpandUpdatedStringList(ctx, plan.Tags, &resp.Diagnostics))
		if resp.Diagnostics.HasError() {
			return
		}

		hasChanges = true
	}

	if !plan.ServerTags.Equal(state.ServerTags) {
		updateReq.ServerTags = new(scwtypes.ExpandUpdatedStringList(ctx, plan.ServerTags, &resp.Diagnostics))
		if resp.Diagnostics.HasError() {
			return
		}

		hasChanges = true
	}

	if !plan.ServerType.Equal(state.ServerType) {
		updateReq.ServerType = new(plan.ServerType.ValueString())
		hasChanges = true
	}

	if !plan.SecurityGroupID.Equal(state.SecurityGroupID) {
		securityGroupID := plan.SecurityGroupID.ValueString()
		updateReq.SecurityGroupID = new(securityGroupID)
		hasChanges = true
	}

	if !plan.PlacementGroupID.Equal(state.PlacementGroupID) {
		placementGroupID := plan.PlacementGroupID.ValueString()
		updateReq.PlacementGroupID = new(placementGroupID)
		hasChanges = true
	}

	if !plan.PublicIPV4Count.Equal(state.PublicIPV4Count) {
		publicIPV4Count := plan.PublicIPV4Count.ValueInt32()
		updateReq.PublicIPV4Count = new(uint32(publicIPV4Count))
		hasChanges = true
	}

	if !plan.PublicIPV6Count.Equal(state.PublicIPV6Count) {
		publicIPV6Count := plan.PublicIPV6Count.ValueInt32()
		updateReq.PublicIPV6Count = new(uint32(publicIPV6Count))
		hasChanges = true
	}

	if !plan.Volumes.Equal(state.Volumes) {
		updateReq.UpdateVolumes = expandVolumes(ctx, plan.Volumes, &resp.Diagnostics).ToUpdateRequest()
		hasChanges = true
	}

	if !plan.PrivateNetworks.Equal(state.PrivateNetworks) {
		updateReq.UpdatePrivateNetworks = expandPrivateNetworks(ctx, plan.PrivateNetworks, &resp.Diagnostics).ToUpdateRequest()
		if resp.Diagnostics.HasError() {
			return
		}

		hasChanges = true
	}

	if !plan.WindowsRdpSSHKeyID.Equal(state.WindowsRdpSSHKeyID) {
		windowsRdpSSHKeyID := plan.WindowsRdpSSHKeyID.ValueString()
		updateReq.WindowsRdpSSHKeyID = new(windowsRdpSSHKeyID)
		hasChanges = true
	}

	if !plan.FilesystemIDs.Equal(state.FilesystemIDs) {
		filesystemIDs := scwtypes.ExpandRawIDSet(ctx, plan.FilesystemIDs, "filesystem_ids", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.FilesystemIDs = new(filesystemIDs)
		hasChanges = true
	}

	if !hasChanges {
		return
	}

	tmpl, err := r.api.UpdateTemplate(updateReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Instance Template", err.Error())

		return
	}

	newState := flattenInstanceTemplate(ctx, tmpl, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, new(newState))...)
}

func (r *InstanceTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Instance Template ID", err.Error())

		return
	}

	err = r.api.DeleteTemplate(&instanceV2.DeleteTemplateRequest{
		Zone:       zone,
		TemplateID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete Instance Template", err.Error())

		return
	}
}

func (r *InstanceTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	zone, id, err := zonal.ParseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse import ID", "Expected format: {zone}/{id}. "+err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), zonal.NewIDString(zone, id))...)
}
