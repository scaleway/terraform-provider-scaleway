package dataorchestrator

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	dataorchestratorSDK "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/workflow_start_action.md
var workflowStartActionDescription string

//go:embed descriptions/workflow_pause_action.md
var workflowPauseActionDescription string

//go:embed descriptions/workflow_resume_action.md
var workflowResumeActionDescription string

//go:embed descriptions/workflow_terminate_action.md
var workflowTerminateActionDescription string

var (
	_ action.Action              = (*StartWorkflowAction)(nil)
	_ action.ActionWithConfigure = (*StartWorkflowAction)(nil)
	_ action.Action              = (*PauseWorkflowRunAction)(nil)
	_ action.ActionWithConfigure = (*PauseWorkflowRunAction)(nil)
	_ action.Action              = (*ResumeWorkflowRunAction)(nil)
	_ action.ActionWithConfigure = (*ResumeWorkflowRunAction)(nil)
	_ action.Action              = (*TerminateWorkflowRunAction)(nil)
	_ action.ActionWithConfigure = (*TerminateWorkflowRunAction)(nil)
)

type workflowActionAPI struct {
	api  *dataorchestratorSDK.API
	meta *meta.Meta
}

func (a *workflowActionAPI) configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.meta = m
	a.api = dataorchestratorSDK.NewAPI(m.ScwClient())
}

// StartWorkflowAction starts a workflow definition.
type StartWorkflowAction struct {
	workflowActionAPI
}

type startWorkflowActionModel struct {
	WorkflowDefinitionID types.String `tfsdk:"workflow_definition_id"`
	Region               types.String `tfsdk:"region"`
	EnvironmentVariables types.Map    `tfsdk:"environment_variables"`
	WaitForCompletion    types.Bool   `tfsdk:"wait_for_completion"`
}

func NewStartWorkflowAction() action.Action {
	return &StartWorkflowAction{}
}

func (a *StartWorkflowAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(ctx, req, resp)
}

func (a *StartWorkflowAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataorchestrator_workflow_start"
}

func (a *StartWorkflowAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: workflowStartActionDescription,
		Description:         workflowStartActionDescription,
		Attributes: map[string]schema.Attribute{
			"workflow_definition_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the workflow definition to start. Plain UUID or regional ID.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Region of the workflow definition. If unset, derived from a regional ID or the provider default region.",
			},
			"environment_variables": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Contextual environment variables for this run.",
			},
			"wait_for_completion": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, wait until the run reaches a terminal state. Defaults to false.",
			},
		},
	}
}

func (a *StartWorkflowAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data startWorkflowActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if a.api == nil {
		resp.Diagnostics.AddError("Unconfigured API", "The action was not properly configured.")

		return
	}

	region, workflowID, err := regional.ResolveID(data.WorkflowDefinitionID.ValueString(), data.Region.ValueString(), a.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())

		return
	}

	startReq := &dataorchestratorSDK.StartWorkflowDefinitionRequest{
		Region:               region,
		WorkflowDefinitionID: workflowID,
	}

	if !data.EnvironmentVariables.IsNull() && !data.EnvironmentVariables.IsUnknown() {
		env := make(map[string]string)
		resp.Diagnostics.Append(data.EnvironmentVariables.ElementsAs(ctx, &env, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(env) > 0 {
			startReq.EnvironmentVariables = &env
		}
	}

	run, err := a.api.StartWorkflowDefinition(startReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to start workflow",
			fmt.Sprintf("Failed to start workflow definition %s: %s", workflowID, err),
		)

		return
	}

	wait := !data.WaitForCompletion.IsNull() && !data.WaitForCompletion.IsUnknown() && data.WaitForCompletion.ValueBool()
	if !wait {
		return
	}

	run, err = a.api.WaitForWorkflowRun(&dataorchestratorSDK.WaitForWorkflowRunRequest{
		Region:        region,
		WorkflowRunID: run.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for workflow run", err.Error())

		return
	}

	if run.Status != dataorchestratorSDK.RunStatusSucceeded {
		msg := fmt.Sprintf("workflow run %s ended with status %q", run.ID, run.Status)
		if run.ErrorMessage != nil && *run.ErrorMessage != "" {
			msg += ": " + *run.ErrorMessage
		}

		resp.Diagnostics.AddError("Workflow run did not succeed", msg)
	}
}

// PauseWorkflowRunAction pauses a workflow run.
type PauseWorkflowRunAction struct {
	workflowActionAPI
}

type workflowRunActionModel struct {
	WorkflowRunID types.String `tfsdk:"workflow_run_id"`
	Region        types.String `tfsdk:"region"`
}

func NewPauseWorkflowRunAction() action.Action {
	return &PauseWorkflowRunAction{}
}

func (a *PauseWorkflowRunAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(ctx, req, resp)
}

func (a *PauseWorkflowRunAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataorchestrator_workflow_pause"
}

func (a *PauseWorkflowRunAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: workflowPauseActionDescription,
		Description:         workflowPauseActionDescription,
		Attributes: map[string]schema.Attribute{
			"workflow_run_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the workflow run to pause.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Region of the workflow run. If unset, derived from a regional ID or the provider default region.",
			},
		},
	}
}

func (a *PauseWorkflowRunAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	invokeWorkflowRunControl(ctx, req, resp, a.api, a.meta, "pause", func(region scw.Region, runID string) error {
		_, err := a.api.PauseWorkflowRun(&dataorchestratorSDK.PauseWorkflowRunRequest{
			Region:        region,
			WorkflowRunID: runID,
		}, scw.WithContext(ctx))

		return err
	})
}

// ResumeWorkflowRunAction resumes a workflow run.
type ResumeWorkflowRunAction struct {
	workflowActionAPI
}

func NewResumeWorkflowRunAction() action.Action {
	return &ResumeWorkflowRunAction{}
}

func (a *ResumeWorkflowRunAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(ctx, req, resp)
}

func (a *ResumeWorkflowRunAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataorchestrator_workflow_resume"
}

func (a *ResumeWorkflowRunAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: workflowResumeActionDescription,
		Description:         workflowResumeActionDescription,
		Attributes: map[string]schema.Attribute{
			"workflow_run_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the workflow run to resume.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Region of the workflow run. If unset, derived from a regional ID or the provider default region.",
			},
		},
	}
}

func (a *ResumeWorkflowRunAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	invokeWorkflowRunControl(ctx, req, resp, a.api, a.meta, "resume", func(region scw.Region, runID string) error {
		_, err := a.api.ResumeWorkflowRun(&dataorchestratorSDK.ResumeWorkflowRunRequest{
			Region:        region,
			WorkflowRunID: runID,
		}, scw.WithContext(ctx))

		return err
	})
}

// TerminateWorkflowRunAction terminates a workflow run.
type TerminateWorkflowRunAction struct {
	workflowActionAPI
}

func NewTerminateWorkflowRunAction() action.Action {
	return &TerminateWorkflowRunAction{}
}

func (a *TerminateWorkflowRunAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.configure(ctx, req, resp)
}

func (a *TerminateWorkflowRunAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataorchestrator_workflow_terminate"
}

func (a *TerminateWorkflowRunAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: workflowTerminateActionDescription,
		Description:         workflowTerminateActionDescription,
		Attributes: map[string]schema.Attribute{
			"workflow_run_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the workflow run to terminate.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Region of the workflow run. If unset, derived from a regional ID or the provider default region.",
			},
		},
	}
}

func (a *TerminateWorkflowRunAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	invokeWorkflowRunControl(ctx, req, resp, a.api, a.meta, "terminate", func(region scw.Region, runID string) error {
		_, err := a.api.TerminateWorkflowRun(&dataorchestratorSDK.TerminateWorkflowRunRequest{
			Region:        region,
			WorkflowRunID: runID,
		}, scw.WithContext(ctx))

		return err
	})
}

func invokeWorkflowRunControl(
	ctx context.Context,
	req action.InvokeRequest,
	resp *action.InvokeResponse,
	api *dataorchestratorSDK.API,
	m *meta.Meta,
	verb string,
	call func(region scw.Region, runID string) error,
) {
	var data workflowRunActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if api == nil {
		resp.Diagnostics.AddError("Unconfigured API", "The action was not properly configured.")

		return
	}

	region, runID, err := regional.ResolveID(data.WorkflowRunID.ValueString(), data.Region.ValueString(), m.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Missing region", err.Error())

		return
	}

	err = call(region, runID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to %s workflow run", verb),
			fmt.Sprintf("Failed to %s workflow run %s: %s", verb, runID, err),
		)
	}
}
