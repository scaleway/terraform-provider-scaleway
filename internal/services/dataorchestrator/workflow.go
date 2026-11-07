package dataorchestrator

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	dataorchestratorSDK "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

var (
	_ resource.Resource                = (*WorkflowResource)(nil)
	_ resource.ResourceWithConfigure   = (*WorkflowResource)(nil)
	_ resource.ResourceWithImportState = (*WorkflowResource)(nil)
	_ resource.ResourceWithIdentity    = (*WorkflowResource)(nil)
)

//go:embed descriptions/workflow.md
var workflowDescription string

func NewWorkflowResource() resource.Resource {
	return &WorkflowResource{}
}

type WorkflowResource struct {
	api  *dataorchestratorSDK.API
	meta *meta.Meta
}

type workflowResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	WorkflowDSL          types.String `tfsdk:"workflow_dsl"`
	VersionName          types.String `tfsdk:"version_name"`
	VersionDescription   types.String `tfsdk:"version_description"`
	VersionID            types.String `tfsdk:"version_id"`
	EnvironmentVariables types.Map    `tfsdk:"environment_variables"`
	WorkflowTimeout      types.String `tfsdk:"workflow_timeout"`
	CronSchedule         types.Object `tfsdk:"cron_schedule"`
	ProjectID            types.String `tfsdk:"project_id"`
	Region               types.String `tfsdk:"region"`
	Srn                  types.String `tfsdk:"srn"`
	CreatedAt            types.String `tfsdk:"created_at"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
}

type workflowResourceIdentityModel = framework.RegionalIdentity

type cronScheduleModel struct {
	Schedule           types.String `tfsdk:"schedule"`
	TimezoneIdentifier types.String `tfsdk:"timezone_identifier"`
	Overlap            types.String `tfsdk:"overlap"`
}

func cronScheduleAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"schedule":            types.StringType,
		"timezone_identifier": types.StringType,
		"overlap":             types.StringType,
	}
}

func (r *WorkflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataorchestrator_workflow"
}

func (r *WorkflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: workflowDescription,
		Description:         workflowDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the workflow definition, in the `{region}/{id}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Name of the workflow definition. If omitted, a random name is generated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the workflow definition.",
			},
			"workflow_dsl": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Open Workflow Specification YAML for the current workflow version.",
			},
			"version_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Name of the current workflow version (3-50 chars: letters, digits, `-`, `_`). Defaults to `v1-0-0` on create.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version_description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the current workflow version.",
			},
			"version_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the current workflow version.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_variables": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Environment variables of the workflow definition.",
			},
			"workflow_timeout": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Timeout of the workflow as a Go duration (e.g. `1h`, `30m`).",
			},
			"cron_schedule": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Optional cron schedule for the workflow.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"schedule": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "UNIX cron expression (e.g. `0 6 * * *`).",
					},
					"timezone_identifier": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Timezone in tz database format (e.g. `Europe/Paris`).",
					},
					"overlap": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Overlap policy: `forbid`, `buffer_one`, `buffer_all`, or `allow_all`.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Project ID. Defaults to the provider project ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": regional.SchemaAttributeComputed("Region of the workflow. Only `fr-par` is currently supported."),
			"srn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Scaleway Resource Name of the workflow definition.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp of the workflow definition.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp of the workflow definition.",
			},
		},
	}
}

func (r *WorkflowResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultRegional()
}

func (r *WorkflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = dataorchestratorSDK.NewAPI(r.meta.ScwClient())
}

func (r *WorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data workflowResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(data.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	versionName := "v1-0-0"
	if !data.VersionName.IsNull() && !data.VersionName.IsUnknown() && data.VersionName.ValueString() != "" {
		versionName = data.VersionName.ValueString()
	}

	dsl := data.WorkflowDSL.ValueString()
	createReq := &dataorchestratorSDK.CreateWorkflowDefinitionRequest{
		Region:      region,
		ProjectID:   projectID,
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		WorkflowVersion: &dataorchestratorSDK.CreateWorkflowDefinitionRequestWorkflowVersionConfig{
			Name:        versionName,
			Description: data.VersionDescription.ValueString(),
			WorkflowDsl: &dsl,
		},
	}

	if !data.EnvironmentVariables.IsNull() && !data.EnvironmentVariables.IsUnknown() {
		env := make(map[string]string)
		resp.Diagnostics.Append(data.EnvironmentVariables.ElementsAs(ctx, &env, false)...)

		if resp.Diagnostics.HasError() {
			return
		}

		createReq.EnvironmentVariables = env
	}

	timeout, diags := expandWorkflowTimeout(data.WorkflowTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq.WorkflowTimeout = timeout

	cron, diags := expandCronScheduleCreate(ctx, data.CronSchedule)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq.CronSchedule = cron

	def, err := r.api.CreateWorkflowDefinition(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create workflow definition", err.Error())

		return
	}

	state := flattenWorkflowDefinition(ctx, def, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(def.Region, def.ID))...)
}

func (r *WorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    workflowResourceModel
		identity workflowResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() && !identity.ID.IsNull() && !identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		id     string
		region scw.Region
		err    error
	)

	if identityAvailable {
		region, id = framework.ResolveRegionalIdentity(identity)
	} else {
		region, id, err = regional.ParseID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse workflow ID", err.Error())

			return
		}
	}

	def, err := r.api.GetWorkflowDefinition(&dataorchestratorSDK.GetWorkflowDefinitionRequest{
		Region:               region,
		WorkflowDefinitionID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read workflow definition", err.Error())

		return
	}

	newState := flattenWorkflowDefinition(ctx, def, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(def.Region, def.ID))...)
}

func (r *WorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  workflowResourceModel
		state workflowResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse workflow ID", err.Error())

		return
	}

	updateDef := &dataorchestratorSDK.UpdateWorkflowDefinitionRequest{
		Region:               region,
		WorkflowDefinitionID: id,
	}
	hasDefChanges := false

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateDef.Name = &name
		hasDefChanges = true
	}

	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateDef.Description = &desc
		hasDefChanges = true
	}

	if !plan.EnvironmentVariables.Equal(state.EnvironmentVariables) {
		env := make(map[string]string)
		if !plan.EnvironmentVariables.IsNull() && !plan.EnvironmentVariables.IsUnknown() {
			resp.Diagnostics.Append(plan.EnvironmentVariables.ElementsAs(ctx, &env, false)...)

			if resp.Diagnostics.HasError() {
				return
			}
		}

		updateDef.EnvironmentVariables = &env
		hasDefChanges = true
	}

	if !plan.WorkflowTimeout.Equal(state.WorkflowTimeout) {
		timeout, diags := expandWorkflowTimeout(plan.WorkflowTimeout)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		updateDef.WorkflowTimeout = timeout
		hasDefChanges = true
	}

	if !plan.CronSchedule.Equal(state.CronSchedule) {
		cron, diags := expandCronScheduleUpdate(ctx, plan.CronSchedule)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		updateDef.CronSchedule = cron
		hasDefChanges = true
	}

	if hasDefChanges {
		_, err = r.api.UpdateWorkflowDefinition(updateDef, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update workflow definition", err.Error())

			return
		}
	}

	versionNeedsUpdate := !plan.WorkflowDSL.Equal(state.WorkflowDSL) ||
		!plan.VersionName.Equal(state.VersionName) ||
		!plan.VersionDescription.Equal(state.VersionDescription)

	if versionNeedsUpdate {
		versionID := state.VersionID.ValueString()
		if versionID == "" {
			resp.Diagnostics.AddError("Missing version_id", "Cannot update workflow version without version_id in state.")

			return
		}

		updateVersion := &dataorchestratorSDK.UpdateWorkflowVersionRequest{
			Region:            region,
			WorkflowVersionID: versionID,
		}

		if !plan.WorkflowDSL.Equal(state.WorkflowDSL) {
			dsl := plan.WorkflowDSL.ValueString()
			updateVersion.WorkflowDsl = &dsl
		}

		if !plan.VersionName.Equal(state.VersionName) {
			name := plan.VersionName.ValueString()
			updateVersion.Name = &name
		}

		if !plan.VersionDescription.Equal(state.VersionDescription) {
			desc := plan.VersionDescription.ValueString()
			updateVersion.Description = &desc
		}

		_, err = r.api.UpdateWorkflowVersion(updateVersion, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update workflow version", err.Error())

			return
		}
	}

	def, err := r.api.GetWorkflowDefinition(&dataorchestratorSDK.GetWorkflowDefinitionRequest{
		Region:               region,
		WorkflowDefinitionID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read workflow definition after update", err.Error())

		return
	}

	newState := flattenWorkflowDefinition(ctx, def, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(def.Region, def.ID))...)
}

func (r *WorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workflowResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse workflow ID", err.Error())

		return
	}

	err = r.api.DeleteWorkflowDefinition(&dataorchestratorSDK.DeleteWorkflowDefinitionRequest{
		Region:               region,
		WorkflowDefinitionID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to delete workflow definition", err.Error())
	}
}

func (r *WorkflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func expandWorkflowTimeout(value types.String) (*scw.Duration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return nil, diags
	}

	d, err := scwtypes.ExpandDuration(value.ValueString())
	if err != nil {
		diags.AddError("Invalid workflow_timeout", err.Error())

		return nil, diags
	}

	if d == nil {
		return nil, diags
	}

	return scw.NewDurationFromTimeDuration(*d), diags
}

func expandCronScheduleCreate(ctx context.Context, obj types.Object) (*dataorchestratorSDK.CreateWorkflowDefinitionRequestCronScheduleConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	var cron cronScheduleModel
	diags.Append(obj.As(ctx, &cron, basetypes.ObjectAsOptions{})...)

	if diags.HasError() {
		return nil, diags
	}

	cfg := &dataorchestratorSDK.CreateWorkflowDefinitionRequestCronScheduleConfig{
		Schedule:           cron.Schedule.ValueString(),
		TimezoneIDentifier: cron.TimezoneIdentifier.ValueString(),
	}

	if !cron.Overlap.IsNull() && !cron.Overlap.IsUnknown() && cron.Overlap.ValueString() != "" {
		overlap := dataorchestratorSDK.ScheduleOverlapPolicy(cron.Overlap.ValueString())
		cfg.Overlap = &overlap
	}

	return cfg, diags
}

func expandCronScheduleUpdate(ctx context.Context, obj types.Object) (*dataorchestratorSDK.UpdateWorkflowDefinitionRequestCronScheduleConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	var cron cronScheduleModel
	diags.Append(obj.As(ctx, &cron, basetypes.ObjectAsOptions{})...)

	if diags.HasError() {
		return nil, diags
	}

	schedule := cron.Schedule.ValueString()
	tz := cron.TimezoneIdentifier.ValueString()
	cfg := &dataorchestratorSDK.UpdateWorkflowDefinitionRequestCronScheduleConfig{
		Schedule:           &schedule,
		TimezoneIDentifier: &tz,
	}

	if !cron.Overlap.IsNull() && !cron.Overlap.IsUnknown() && cron.Overlap.ValueString() != "" {
		overlap := dataorchestratorSDK.ScheduleOverlapPolicy(cron.Overlap.ValueString())
		cfg.Overlap = &overlap
	}

	return cfg, diags
}

func flattenWorkflowDefinition(ctx context.Context, def *dataorchestratorSDK.WorkflowDefinition, diags *diag.Diagnostics) workflowResourceModel {
	model := workflowResourceModel{
		ID:          types.StringValue(regional.NewIDString(def.Region, def.ID)),
		Name:        types.StringValue(def.Name),
		Description: types.StringValue(def.Description),
		ProjectID:   types.StringValue(def.ProjectID),
		Region:      types.StringValue(def.Region.String()),
		Srn:         types.StringValue(def.Srn),
	}

	if def.CreatedAt != nil {
		model.CreatedAt = types.StringValue(def.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringNull()
	}

	if def.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(def.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	if len(def.EnvironmentVariables) == 0 {
		model.EnvironmentVariables = types.MapNull(types.StringType)
	} else {
		m, mapDiags := types.MapValueFrom(ctx, types.StringType, def.EnvironmentVariables)
		diags.Append(mapDiags...)

		model.EnvironmentVariables = m
	}

	if def.WorkflowTimeout != nil {
		td := def.WorkflowTimeout.ToTimeDuration()
		if td != nil {
			model.WorkflowTimeout = types.StringValue(td.String())
		} else {
			model.WorkflowTimeout = types.StringNull()
		}
	} else {
		model.WorkflowTimeout = types.StringNull()
	}

	if def.CronSchedule != nil {
		cron := cronScheduleModel{
			Schedule:           types.StringValue(def.CronSchedule.Schedule),
			TimezoneIdentifier: types.StringValue(def.CronSchedule.TimezoneIDentifier),
			Overlap:            types.StringValue(def.CronSchedule.Overlap.String()),
		}
		obj, objDiags := types.ObjectValueFrom(ctx, cronScheduleAttrTypes(), cron)
		diags.Append(objDiags...)

		model.CronSchedule = obj
	} else {
		model.CronSchedule = types.ObjectNull(cronScheduleAttrTypes())
	}

	version, err := currentWorkflowVersion(def)
	if err != nil {
		diags.AddError("Failed to resolve current workflow version", err.Error())

		model.WorkflowDSL = types.StringNull()
		model.VersionName = types.StringNull()
		model.VersionDescription = types.StringNull()
		model.VersionID = types.StringNull()

		return model
	}

	model.WorkflowDSL = types.StringValue(version.WorkflowDsl)
	model.VersionName = types.StringValue(version.Name)
	model.VersionDescription = types.StringValue(version.Description)
	model.VersionID = types.StringValue(version.ID)

	return model
}
