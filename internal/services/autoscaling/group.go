package autoscaling

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	autoscaling "github.com/scaleway/scaleway-sdk-go/api/autoscaling/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*AutoScalingGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*AutoScalingGroupResource)(nil)
	_ resource.ResourceWithImportState = (*AutoScalingGroupResource)(nil)
)

func NewAutoScalingGroupResource() resource.Resource {
	return &AutoScalingGroupResource{}
}

type AutoScalingGroupResource struct {
	api  *autoscaling.API
	meta *meta.Meta
}

type autoScalingGroupResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	TemplateID                types.String `tfsdk:"template_id"`
	ScalingPolicy             types.Object `tfsdk:"scaling_policy"`
	LoadBalancerConfiguration types.Object `tfsdk:"load_balancer_configuration"`
	Tags                      types.List   `tfsdk:"tags"`
	Zone                      types.String `tfsdk:"zone"`
	Status                    types.String `tfsdk:"status"`
	ProjectID                 types.String `tfsdk:"project_id"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
	CreatedAt                 types.String `tfsdk:"created_at"`
}

func (r *AutoScalingGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_autoscaling_group"
}

func (r *AutoScalingGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway AutoScaling Group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the AutoScaling Group, in the `{zone}/{id}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the AutoScaling Group. If not provided, a random name will be generated.",
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tags associated with the AutoScaling Group.",
			},
			"template_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the Instance Template used to create instances in this group.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
				},
			},
			"scaling_policy": schema.SingleNestedAttribute{
				Attributes:          scalingPolicyAttributes(),
				Required:            true,
				MarkdownDescription: "The scaling policy configuration.",
				Validators: []validator.Object{
					scalingPolicyTargetValidator{},
				},
			},
			"load_balancer_configuration": schema.SingleNestedAttribute{
				Attributes:          loadBalancerConfigAttributes(),
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The load balancer configuration.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current status of the AutoScaling Group.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creation timestamp of the AutoScaling Group.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The last update timestamp of the AutoScaling Group.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the AutoScaling Group belongs to. Defaults to the provider's project ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The zone the AutoScaling Group is in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func scalingPolicyAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"minimum_size": schema.Int32Attribute{
			Required:            true,
			MarkdownDescription: "The minimum number of instances in the group.",
		},
		"maximum_size": schema.Int32Attribute{
			Required:            true,
			MarkdownDescription: "The maximum number of instances in the group.",
		},
		"scale_in_cooldown": schema.StringAttribute{
			CustomType:          timetypes.GoDurationType{},
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("1m"),
			MarkdownDescription: "The cooldown duration after a scale-in event.",
		},
		"scale_out_cooldown": schema.StringAttribute{
			CustomType:          timetypes.GoDurationType{},
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("1m"),
			MarkdownDescription: "The cooldown duration after a scale-out event.",
		},
		"scale_in_step": schema.Int32Attribute{
			Optional:            true,
			Computed:            true,
			Default:             int32default.StaticInt32(1),
			MarkdownDescription: "The number of instances to remove during scale-in event.",
		},
		"scale_out_step": schema.Int32Attribute{
			Optional:            true,
			Computed:            true,
			Default:             int32default.StaticInt32(1),
			MarkdownDescription: "The number of instances to add during scale-out event.",
		},
		"fixed_size": schema.Int32Attribute{
			Optional:            true,
			MarkdownDescription: "The fixed number of instances for the group.",
		},
		"cpu_target": schema.Int32Attribute{
			Optional:            true,
			MarkdownDescription: "The target CPU utilization percentage to trigger scaling events.",
		},
		"memory_target": schema.Int32Attribute{
			Optional:            true,
			MarkdownDescription: "The target memory utilization percentage to trigger scaling events.",
		},
	}
}

func loadBalancerConfigAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"load_balancer_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The ID of the load balancer.",
			Validators: []validator.String{
				verify.IsStringUUIDOrUUIDWithZone(),
			},
		},
		"backends": schema.ListNestedAttribute{
			Required:            true,
			MarkdownDescription: "The list of load balancer backend configurations.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"backend_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The ID of the load balancer backend.",
						Validators: []validator.String{
							verify.IsStringUUIDOrUUIDWithZone(),
						},
					},
					"address_family": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The IP address family (IPv4 or IPv6).",
						Validators: []validator.String{
							verify.ValidateEnumFramework[autoscaling.GroupLoadBalancerConfigurationBackendAddressFamily](),
						},
					},
					"private_network_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The ID of the private network.",
						Validators: []validator.String{
							verify.IsStringUUIDOrUUIDWithRegion(),
						},
					},
				},
			},
		},
		"auto_healing": schema.SingleNestedAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The auto-healing configuration.",
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether auto-healing is enabled.",
				},
				"grace_period": schema.StringAttribute{
					CustomType:          timetypes.GoDurationType{},
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "The grace period for health checks.",
				},
			},
		},
	}
}

func backendObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"backend_id":         types.StringType,
			"address_family":     types.StringType,
			"private_network_id": types.StringType,
		},
	}
}

func autoHealingObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled":      types.BoolType,
			"grace_period": timetypes.GoDurationType{},
		},
	}
}

func (r *AutoScalingGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = autoscaling.NewAPI(r.meta.ScwClient())
}

func (r *AutoScalingGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data autoScalingGroupResourceModel
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

	templateID := scwtypes.ExpandRawID(data.TemplateID, "template_id", &resp.Diagnostics)
	if resp.Diagnostics.HasError() || templateID == nil {
		return
	}

	createReq := &autoscaling.CreateGroupRequest{
		Zone:                          zone,
		ProjectID:                     projectID,
		Name:                          scwtypes.ExpandOrGenerateString(data.Name.ValueString(), "tf-asg"),
		TemplateID:                    *templateID,
		Tags:                          scwtypes.ExpandStringList(ctx, data.Tags, &resp.Diagnostics),
		ScalingPolicySpec:             expandScalingPolicy(data.ScalingPolicy),
		LoadBalancerConfigurationSpec: expandLoadBalancerConfiguration(ctx, data.LoadBalancerConfiguration, &resp.Diagnostics),
	}

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.api.CreateGroup(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Autoscaling Group", err.Error())

		return
	}

	if ok, warning := checkZone(ctx, r, zone, projectID, *templateID, group.ID); !ok {
		resp.Diagnostics.Append(warning)
	}

	state := flattenGroup(ctx, group, zone, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func checkZone(ctx context.Context, r *AutoScalingGroupResource, zone scw.Zone, projectID, templateID, groupID string) (bool, diag.Diagnostic) {
	pageToken := new("")

	for {
		groupSummaries, err := r.api.ListGroups(&autoscaling.ListGroupsRequest{
			Zone:       zone,
			ProjectID:  projectID,
			TemplateID: &templateID,
			PageToken:  pageToken,
		}, scw.WithContext(ctx))
		if err != nil {
			return false, diag.NewWarningDiagnostic(
				"Could not verify zone for Autoscaling Group "+groupID,
				"Failed to list groups: "+err.Error(),
			)
		}

		for _, gs := range groupSummaries.GroupSummaries {
			if gs.ID == groupID {
				if gs.Zone == zone {
					return true, nil
				}

				// Should not ever return here
				return false, diag.NewWarningDiagnostic(
					"unexpected zone for group",
					fmt.Sprintf("expected %s, got %s", zone, gs.Zone),
				)
			}
		}

		if groupSummaries.NextPageToken == nil {
			break
		}

		pageToken = groupSummaries.NextPageToken
	}

	return false, diag.NewWarningDiagnostic(
		"Could not verify zone for Autoscaling Group "+groupID,
		"The zone from the config will be used during reading operations, which can cause some unexpected behaviors.",
	)
}

func flattenGroup(ctx context.Context, group *autoscaling.Group, zone scw.Zone, reference any, diags *diag.Diagnostics) any {
	model := autoScalingGroupResourceModel{
		ID:        types.StringValue(zonal.NewIDString(zone, group.ID)),
		Name:      types.StringValue(group.Name),
		Zone:      types.StringValue(zone.String()),
		ProjectID: types.StringValue(group.ProjectID),
		Status:    types.StringValue(group.Status.String()),
	}

	if scwtypes.IDUsesZonedFormat(ctx, reference, path.Root("template_id"), diags) {
		model.TemplateID = types.StringValue(zonal.NewIDString(zone, group.TemplateID))
	} else {
		model.TemplateID = types.StringValue(group.TemplateID)
	}

	tagList, d := scwtypes.FlattenStringList(ctx, "tags", group.Tags, reference)
	diags.Append(d...)

	model.Tags = tagList

	scalingPolicy, d := flattenScalingPolicy(group.ScalingPolicy)
	diags.Append(d...)

	model.ScalingPolicy = scalingPolicy

	loadBalancerConfiguration, d := flattenLoadBalancerConfiguration(ctx, group.LoadBalancerConfiguration, zone, reference)
	diags.Append(d...)

	model.LoadBalancerConfiguration = loadBalancerConfiguration

	if group.CreatedAt != nil {
		model.CreatedAt = types.StringValue(group.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringNull()
	}

	if group.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(group.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	return model
}

func (r *AutoScalingGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state autoScalingGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Autoscaling Group ID", err.Error())

		return
	}

	group, err := r.api.GetGroup(&autoscaling.GetGroupRequest{
		Zone:    zone,
		GroupID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read Autoscaling Group", err.Error())

		return
	}

	newState := flattenGroup(ctx, group, zone, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *AutoScalingGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  autoScalingGroupResourceModel
		state autoScalingGroupResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Autoscaling Group ID", err.Error())

		return
	}

	updateReq := &autoscaling.UpdateGroupRequest{
		Zone:    zone,
		GroupID: id,
	}
	hasChanges := false

	if !plan.Name.Equal(state.Name) {
		updateReq.Name = new(plan.Name.ValueString())
		hasChanges = true
	}

	if !plan.Tags.Equal(state.Tags) {
		updateReq.Tags = new(scwtypes.ExpandUpdatedStringList(ctx, plan.Tags, &resp.Diagnostics))
		hasChanges = true
	}

	if !plan.TemplateID.Equal(state.TemplateID) {
		updateReq.TemplateID = scwtypes.ExpandRawID(plan.TemplateID, "template_id", &resp.Diagnostics)
		hasChanges = true
	}

	if !plan.ScalingPolicy.Equal(state.ScalingPolicy) {
		updateReq.ScalingPolicySpec = expandScalingPolicy(plan.ScalingPolicy)
		hasChanges = true
	}

	if !plan.LoadBalancerConfiguration.Equal(state.LoadBalancerConfiguration) {
		updateReq.LoadBalancerConfigurationSpec = expandLoadBalancerConfiguration(ctx, plan.LoadBalancerConfiguration, &resp.Diagnostics)
		hasChanges = true
	}

	if !hasChanges || resp.Diagnostics.HasError() {
		return
	}

	group, err := r.api.UpdateGroup(updateReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Autoscaling Group", err.Error())

		return
	}

	newState := flattenGroup(ctx, group, zone, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *AutoScalingGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state autoScalingGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse AutoScaling Group ID", err.Error())

		return
	}

	_, err = r.api.DeleteGroup(&autoscaling.DeleteGroupRequest{
		Zone:    zone,
		GroupID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete AutoScaling Group", err.Error())

		return
	}

	_, err = r.api.WaitForGroup(&autoscaling.WaitForGroupRequest{
		Zone:    zone,
		GroupID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to wait for AutoScaling Group after deletion", err.Error())

		return
	}
}

func (r *AutoScalingGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	zone, id, err := zonal.ParseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse import ID", "Expected format: {zone}/{id}. "+err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), zonal.NewIDString(zone, id))...)
}
