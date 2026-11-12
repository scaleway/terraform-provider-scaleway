package dataorchestrator

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	dataorchestratorSDK "github.com/scaleway/scaleway-sdk-go/api/dataorchestrator/v1alpha2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*WorkflowRunDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*WorkflowRunDataSource)(nil)
)

//go:embed descriptions/workflow_run_data_source.md
var workflowRunDataSourceDescription string

func NewWorkflowRunDataSource() datasource.DataSource {
	return &WorkflowRunDataSource{}
}

type WorkflowRunDataSource struct {
	api  *dataorchestratorSDK.API
	meta *meta.Meta
}

type workflowRunDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	RunID                types.String `tfsdk:"run_id"`
	WorkflowDefinitionID types.String `tfsdk:"workflow_definition_id"`
	WorkflowVersionID    types.String `tfsdk:"workflow_version_id"`
	Status               types.String `tfsdk:"status"`
	TriggerType          types.String `tfsdk:"trigger_type"`
	WorkflowDSL          types.String `tfsdk:"workflow_dsl"`
	Output               types.String `tfsdk:"output"`
	ErrorMessage         types.String `tfsdk:"error_message"`
	ProjectID            types.String `tfsdk:"project_id"`
	Region               types.String `tfsdk:"region"`
	Srn                  types.String `tfsdk:"srn"`
	CreatedAt            types.String `tfsdk:"created_at"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
	TerminatedAt         types.String `tfsdk:"terminated_at"`
}

func (d *WorkflowRunDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataorchestrator_workflow_run"
}

func (d *WorkflowRunDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: workflowRunDataSourceDescription,
		Description:         workflowRunDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the workflow run, in the `{region}/{id}` format.",
			},
			"run_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the workflow run. Can be a regional ID.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"workflow_definition_id": schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the workflow definition."},
			"workflow_version_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the workflow version that was run."},
			"status":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Status of the workflow run."},
			"trigger_type":           schema.StringAttribute{Computed: true, MarkdownDescription: "Trigger type of the workflow run."},
			"workflow_dsl":           schema.StringAttribute{Computed: true, MarkdownDescription: "DSL snapshot used for this run."},
			"output":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Output of the workflow run, if any."},
			"error_message":          schema.StringAttribute{Computed: true, MarkdownDescription: "Error message of the workflow run, if any."},
			"project_id":             schema.StringAttribute{Computed: true, MarkdownDescription: "Project ID."},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Region of the workflow run.",
			},
			"srn":           schema.StringAttribute{Computed: true, MarkdownDescription: "Scaleway Resource Name."},
			"created_at":    schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"updated_at":    schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
			"terminated_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Termination timestamp, if any."},
		},
	}
}

func (d *WorkflowRunDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T.", req.ProviderData),
		)

		return
	}

	d.meta = m
	d.api = dataorchestratorSDK.NewAPI(m.ScwClient())
}

func (d *WorkflowRunDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data workflowRunDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, runID, err := regional.ResolveID(data.RunID.ValueString(), data.Region.ValueString(), d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	run, err := d.api.GetWorkflowRun(&dataorchestratorSDK.GetWorkflowRunRequest{
		Region:        region,
		WorkflowRunID: runID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get workflow run", err.Error())

		return
	}

	out := workflowRunDataSourceModel{
		ID:                   types.StringValue(regional.NewIDString(run.Region, run.ID)),
		RunID:                types.StringValue(run.ID),
		WorkflowDefinitionID: types.StringValue(run.WorkflowDefinitionID),
		WorkflowVersionID:    types.StringValue(run.WorkflowVersionID),
		Status:               types.StringValue(run.Status.String()),
		TriggerType:          types.StringValue(run.TriggerType.String()),
		WorkflowDSL:          types.StringValue(run.WorkflowDsl),
		ProjectID:            types.StringValue(run.ProjectID),
		Region:               types.StringValue(run.Region.String()),
		Srn:                  types.StringValue(run.Srn),
	}

	if run.Output != nil {
		out.Output = types.StringValue(*run.Output)
	} else {
		out.Output = types.StringNull()
	}

	if run.ErrorMessage != nil {
		out.ErrorMessage = types.StringValue(*run.ErrorMessage)
	} else {
		out.ErrorMessage = types.StringNull()
	}

	if run.CreatedAt != nil {
		out.CreatedAt = types.StringValue(run.CreatedAt.Format(time.RFC3339))
	}

	if run.UpdatedAt != nil {
		out.UpdatedAt = types.StringValue(run.UpdatedAt.Format(time.RFC3339))
	}

	if run.TerminatedAt != nil {
		out.TerminatedAt = types.StringValue(run.TerminatedAt.Format(time.RFC3339))
	} else {
		out.TerminatedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
