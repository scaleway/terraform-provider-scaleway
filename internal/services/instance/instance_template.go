package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	instanceV2 "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	types2 "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
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
	ProjectID          types.String `tfsdk:"project_id"`
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Tags               types.List   `tfsdk:"tags"`
	ServerTags         types.List   `tfsdk:"server_tags"`
	ServerType         types.String `tfsdk:"server_type"`
	SecurityGroupID    types.String `tfsdk:"security_group_id"`
	PlacementGroupID   types.String `tfsdk:"placement_group_id"`
	PublicIPV4Count    types.Int32  `tfsdk:"public_ip_v4_count"`
	PublicIPV6Count    types.Int32  `tfsdk:"public_ip_v6_count"`
	Volumes            types.List   `tfsdk:"volumes"`
	PrivateNetworks    types.List   `tfsdk:"private_networks"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
	WindowsRdpSSHKeyID types.String `tfsdk:"windows_rdp_ssh_key_id"`
	FilesystemIDs      types.List   `tfsdk:"filesystem_ids"`
	Zone               types.String `tfsdk:"zone"`
}

func (r *InstanceTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_template"
}

func (r *InstanceTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway Instance Template.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the Instance Template belongs to. Defaults to the provider's project ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
				MarkdownDescription: "The name of the Instance Template. If not provided, a random name is generated.", // TODO: is it generated though ?
				//PlanModifiers: []planmodifier.String{
				//	stringplanmodifier.UseStateForUnknown(),
				//},
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
				MarkdownDescription: "The tags that will be assigned to the servers created using the Instance Template.", // TODO: update desc + what happens if undefined ??
			},
			"security_group_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the security group to attach to the servers created using the Instance Template.",
			},
			"placement_group_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the placement group to attach to the servers created using the Instance Template.",
			},
			"public_ip_v4_count": schema.Int32Attribute{
				Required:            true, // TODO: or optional + default=0 ?
				MarkdownDescription: "The number of public IPv4 to attach to the servers created using the Instance Template.",
			},
			"public_ip_v6_count": schema.Int32Attribute{
				Required:            true, // TODO: or optional + default=0 ?
				MarkdownDescription: "The number of public IPv6 to attach to the servers created using the Instance Template.",
			},
			"volumes": schema.ListNestedAttribute{ // schema.ListNestedBlock // TODO: name it 'volume' instead ??
				Optional:            true,
				MarkdownDescription: "The specs of the volumes of the servers created using the Instance Template.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"volume_type": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(instanceV2.CreateServerRequestServerVolumeVolumeTypeUnknownVolumeType.String()),
							MarkdownDescription: "The type of volume, defaults to unknown_volume_type.",
						},
						"name": schema.StringAttribute{
							Optional: true,
							Computed: true,
							// Default ?? generate ??
						},
						"tags": schema.ListAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "The tags associated with the volume.",
						},
						"size": schema.Int64Attribute{
							Optional: true, // TODO: really ??
						},
						"base_snapshot_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("image_label")),
							},
						},
						"image_label": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("base_snapshot_id")),
							},
						},
						"perf_iops": schema.Int32Attribute{
							Optional: true,
						},
					},
				},
			},
			"private_networks": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "The IDs of the private networks to attach to the servers created using the Instance Template.",
				ElementType:         types.StringType,
				//PlanModifiers: []planmodifier.List{
				//	dsf.IDListLocalizer(),
				//},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creation timestamp of the Instance template.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The last update timestamp of the Instance template.",
			},
			"windows_rdp_ssh_key_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the IAM SSH key used to encrypt the initial admin password on a Windows server. This will be repeated on all servers created using the Instance Template.",
			},
			"filesystem_ids": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "The IDs of the filesystems to attach to the servers created using the Instance Template.",
				ElementType:         types.StringType,
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
		Name:            types2.ExpandOrGenerateString(data.Name.ValueString(), "tf-tmpl"),
		ServerType:      data.ServerType.ValueString(),
		PublicIPV4Count: uint32(data.PublicIPV4Count.ValueInt32()),
		PublicIPV6Count: uint32(data.PublicIPV6Count.ValueInt32()),
	}

	createReq.Tags = expandStringList(ctx, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.ServerTags = expandStringList(ctx, data.ServerTags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.SecurityGroupID = expandRawID(data.SecurityGroupID, "security_group_id", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.PlacementGroupID = expandRawID(data.PlacementGroupID, "placement_group_id", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq.WindowsRdpSSHKeyID = expandRawID(data.WindowsRdpSSHKeyID, "windows_rdp_ssh_key_id", &resp.Diagnostics)
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

	createReq.FilesystemIDs = expandRawIDList(ctx, data.FilesystemIDs, "filesystem_ids", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.api.CreateTemplate(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Instance Template", err.Error())

		return
	}

	state := flattenInstanceTemplate(ctx, tmpl, req, &resp.Diagnostics)
	tmpDiags := resp.State.Set(ctx, new(state))
	resp.Diagnostics.Append(tmpDiags...) // TODO: collapse in oneliner
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

	tagList, d := flattenStringList(ctx, tmpl.Tags)
	diags.Append(d...)

	model.Tags = tagList

	serverTagList, d := flattenStringList(ctx, tmpl.ServerTags)
	diags.Append(d...)

	model.ServerTags = serverTagList

	if tmpl.SecurityGroupID != nil {
		if idUsesZonedFormat(ctx, reference, "security_group_id", diags) { // zonal.ExpandID(confSG).Zone != "" {
			model.SecurityGroupID = types.StringValue(zonal.NewIDString(tmpl.Zone, *tmpl.SecurityGroupID))
		} else {
			model.SecurityGroupID = types.StringValue(*tmpl.SecurityGroupID)
		}
	} else {
		model.SecurityGroupID = types.StringNull()
	}

	if tmpl.PlacementGroupID != nil {
		if idUsesZonedFormat(ctx, reference, "placement_group_id", diags) {
			model.PlacementGroupID = types.StringValue(zonal.NewIDString(tmpl.Zone, *tmpl.PlacementGroupID))
		} else {
			model.PlacementGroupID = types.StringValue(*tmpl.PlacementGroupID)
		}
	} else {
		model.PlacementGroupID = types.StringNull()
	}

	volumesList, d := flattenVolumes(ctx, tmpl.Volumes)
	diags.Append(d...)

	model.Volumes = volumesList

	privateNetworkIDsList, d := flattenPrivateNetworks(ctx, tmpl.PrivateNetworks, tmpl.Zone)
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

	region, err := tmpl.Zone.Region()
	if err != nil {
		// TODO: do stuff
	}

	filesystemIDsList, d := flattenLocalizedIDList(ctx, tmpl.FilesystemIDs, region.String())
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
		updateReq.Tags = new(expandUpdatedStringList(ctx, plan.Tags, &resp.Diagnostics))
		if resp.Diagnostics.HasError() {
			return
		}

		hasChanges = true
	}

	if !plan.ServerTags.Equal(state.ServerTags) {
		updateReq.ServerTags = new(expandUpdatedStringList(ctx, plan.ServerTags, &resp.Diagnostics))
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
		filesystemIDs := expandRawIDList(ctx, plan.FilesystemIDs, "filesystem_ids", &resp.Diagnostics)
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
