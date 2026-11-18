package dataorchestrator

import (
	"context"
	_ "embed"
	"fmt"

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
	_ datasource.DataSource              = (*WorkflowDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*WorkflowDataSource)(nil)
)

//go:embed descriptions/workflow_data_source.md
var workflowDataSourceDescription string

func NewWorkflowDataSource() datasource.DataSource {
	return &WorkflowDataSource{}
}

type WorkflowDataSource struct {
	api  *dataorchestratorSDK.API
	meta *meta.Meta
}

type workflowDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	WorkflowID           types.String `tfsdk:"workflow_id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	WorkflowDSL          types.String `tfsdk:"workflow_dsl"`
	VersionName          types.String `tfsdk:"version_name"`
	VersionDescription   types.String `tfsdk:"version_description"`
	VersionID            types.String `tfsdk:"version_id"`
	EnvironmentVariables types.Map    `tfsdk:"environment_variables"`
	CronSchedule         types.Object `tfsdk:"cron_schedule"`
	ProjectID            types.String `tfsdk:"project_id"`
	Region               types.String `tfsdk:"region"`
	Srn                  types.String `tfsdk:"srn"`
	CreatedAt            types.String `tfsdk:"created_at"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
}

func (d *WorkflowDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataorchestrator_workflow"
}

func (d *WorkflowDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: workflowDataSourceDescription,
		Description:         workflowDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the workflow definition, in the `{region}/{id}` format.",
			},
			"workflow_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the workflow definition. Can be a regional ID.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"name":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Name of the workflow definition."},
			"description":           schema.StringAttribute{Computed: true, MarkdownDescription: "Description of the workflow definition."},
			"workflow_dsl":          schema.StringAttribute{Computed: true, MarkdownDescription: "Current workflow version DSL (YAML)."},
			"version_name":          schema.StringAttribute{Computed: true, MarkdownDescription: "Name of the current workflow version."},
			"version_description":   schema.StringAttribute{Computed: true, MarkdownDescription: "Description of the current workflow version."},
			"version_id":            schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the current workflow version."},
			"environment_variables": schema.MapAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Environment variables of the workflow."},
			"cron_schedule": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Cron schedule of the workflow, if any.",
				Attributes: map[string]schema.Attribute{
					"schedule":            schema.StringAttribute{Computed: true, MarkdownDescription: "UNIX cron expression."},
					"timezone_identifier": schema.StringAttribute{Computed: true, MarkdownDescription: "Timezone identifier."},
					"overlap":             schema.StringAttribute{Computed: true, MarkdownDescription: "Overlap policy."},
				},
			},
			"project_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Project ID."},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Region of the workflow.",
			},
			"srn":        schema.StringAttribute{Computed: true, MarkdownDescription: "Scaleway Resource Name."},
			"created_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"updated_at": schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
		},
	}
}

func (d *WorkflowDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkflowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data workflowDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, workflowID, err := regional.ResolveID(data.WorkflowID.ValueString(), data.Region.ValueString(), d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	def, err := d.api.GetWorkflowDefinition(&dataorchestratorSDK.GetWorkflowDefinitionRequest{
		Region:               region,
		WorkflowDefinitionID: workflowID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get workflow definition", err.Error())

		return
	}

	flat := flattenWorkflowDefinition(ctx, def, &resp.Diagnostics)
	out := workflowDataSourceModel{
		ID:                   flat.ID,
		WorkflowID:           types.StringValue(def.ID),
		Name:                 flat.Name,
		Description:          flat.Description,
		WorkflowDSL:          flat.WorkflowDSL,
		VersionName:          flat.VersionName,
		VersionDescription:   flat.VersionDescription,
		VersionID:            flat.VersionID,
		EnvironmentVariables: flat.EnvironmentVariables,
		CronSchedule:         flat.CronSchedule,
		ProjectID:            flat.ProjectID,
		Region:               flat.Region,
		Srn:                  flat.Srn,
		CreatedAt:            flat.CreatedAt,
		UpdatedAt:            flat.UpdatedAt,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
