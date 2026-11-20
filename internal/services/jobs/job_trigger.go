package jobs

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	jobs "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*TriggerResource)(nil)
	_ resource.ResourceWithConfigure   = (*TriggerResource)(nil)
	_ resource.ResourceWithImportState = (*TriggerResource)(nil)
)

func NewTriggerResource() resource.Resource {
	return &TriggerResource{}
}

type TriggerResource struct {
	jobsAPI *jobs.API
	meta    *meta.Meta
}

type triggerResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	JobDefinitionID types.String `tfsdk:"job_definition_id"`
	CronConfig      types.Object `tfsdk:"cron_config"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	Region          types.String `tfsdk:"region"`
	SRN             types.String `tfsdk:"srn"`
}

func (r *TriggerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_trigger"
}

//go:embed descriptions/job_trigger_resource.md
var triggerResourceDescription string

func (r *TriggerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: triggerResourceDescription,
		Description:         triggerResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the trigger.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the trigger. If not provided, a random name will be generated.",
			},
			"job_definition_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the job definition to trigger.",
				Required:            true,
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"cron_config": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "The configuration of the CRON trigger.",
				Attributes: map[string]schema.Attribute{
					"schedule": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The CRON schedule in UNIX format.",
					},
					"timezone": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The timezone for the CRON schedule.",
					},
					"startup_command": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "The startup command that will be used by the triggered job.",
					},
					"args": schema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "The arguments passed to the startup command used by the triggered job.",
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creation time of the trigger.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The last update time of the trigger.",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region in which the trigger exists.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"srn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Scaleway Resource Name (SRN) of the trigger.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *TriggerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.jobsAPI = jobs.NewAPI(r.meta.ScwClient())
}

func (r *TriggerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data triggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(data.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	jobDefinitionID := scwtypes.ExpandRawID(data.JobDefinitionID, "job_definition_id", &resp.Diagnostics)
	if resp.Diagnostics.HasError() || jobDefinitionID == nil {
		return
	}

	createReq := &jobs.CreateTriggerRequest{
		Region:          region,
		JobDefinitionID: *jobDefinitionID,
		Name:            scwtypes.ExpandOrGenerateString(data.Name.ValueString(), "tf-job-trig"),
		CronConfig:      expandCronConfig(ctx, data.CronConfig, &resp.Diagnostics).ToCreateRequest(),
	}

	trigger, err := r.jobsAPI.CreateTrigger(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Job Trigger", err.Error())

		return
	}

	state := flattenTrigger(ctx, trigger, region, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenTrigger(ctx context.Context, trigger *jobs.Trigger, region scw.Region, reference any, diags *diag.Diagnostics) any {
	model := triggerResourceModel{
		ID:     types.StringValue(regional.NewIDString(region, trigger.ID)),
		Name:   types.StringValue(trigger.Name),
		Region: types.StringValue(region.String()),
		SRN:    types.StringValue(trigger.Srn),
	}

	if scwtypes.IDUsesRegionalFormat(ctx, reference, path.Root("job_definition_id"), diags) {
		model.JobDefinitionID = types.StringValue(regional.NewIDString(region, trigger.JobDefinitionID))
	} else {
		model.JobDefinitionID = types.StringValue(trigger.JobDefinitionID)
	}

	model.CronConfig = flattenCronConfig(ctx, trigger.CronConfig, reference, diags)

	if trigger.CreatedAt != nil {
		model.CreatedAt = types.StringValue(trigger.CreatedAt.Format(time.RFC3339))
	}

	if trigger.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(trigger.UpdatedAt.Format(time.RFC3339))
	}

	return model
}

func (r *TriggerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state triggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Job Trigger ID", err.Error())

		return
	}

	trigger, err := r.jobsAPI.GetTrigger(&jobs.GetTriggerRequest{
		Region:    region,
		TriggerID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read Job Trigger", err.Error())

		return
	}

	newState := flattenTrigger(ctx, trigger, region, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *TriggerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  triggerResourceModel
		state triggerResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Job Trigger ID", err.Error())

		return
	}

	updateReq := &jobs.UpdateTriggerRequest{
		Region:    region,
		TriggerID: id,
	}
	hasChanges := false

	if !plan.Name.Equal(state.Name) {
		updateReq.Name = new(plan.Name.ValueString())
		hasChanges = true
	}

	if !plan.CronConfig.Equal(state.CronConfig) {
		updateReq.CronConfig = expandCronConfig(ctx, plan.CronConfig, &resp.Diagnostics).ToUpdateRequest()
		hasChanges = true
	}

	if !hasChanges || resp.Diagnostics.HasError() {
		return
	}

	trigger, err := r.jobsAPI.UpdateTrigger(updateReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Job Trigger", err.Error())

		return
	}

	newState := flattenTrigger(ctx, trigger, region, req, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *TriggerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state triggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Job Trigger ID", err.Error())

		return
	}

	err = r.jobsAPI.DeleteTrigger(&jobs.DeleteTriggerRequest{
		Region:    region,
		TriggerID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete Job Trigger", err.Error())

		return
	}
}

func (r *TriggerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	region, id, err := regional.ParseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse import ID", "Expected format: {region}/{id}. "+err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), regional.NewIDString(region, id))...)
}
