package jobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	jobs "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*StartJobDefinitionAction)(nil)
	_ action.ActionWithConfigure = (*StartJobDefinitionAction)(nil)
)

type StartJobDefinitionAction struct {
	jobsAPI *jobs.API
	meta    *meta.Meta
}

func (a *StartJobDefinitionAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

	client := m.ScwClient()
	a.jobsAPI = jobs.NewAPI(client)
	a.meta = m
}

func (a *StartJobDefinitionAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_definition_start_action"
}

type StartJobDefinitionActionModel struct {
	JobDefinitionID      types.String `tfsdk:"job_definition_id"`
	Region               types.String `tfsdk:"region"`
	Command              types.String `tfsdk:"command"`
	EnvironmentVariables types.Map    `tfsdk:"environment_variables"`
	Replicas             types.Int64  `tfsdk:"replicas"`
}

func NewStartJobDefinitionAction() action.Action {
	return &StartJobDefinitionAction{}
}

func (a *StartJobDefinitionAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"job_definition_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the job definition to start. Can be a plain UUID or a regional ID.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Region of the job definition. If not set, the region is derived from the job_definition_id when possible or from the provider configuration.",
			},
			"command": schema.StringAttribute{
				Optional:    true,
				Description: "Contextual startup command for this specific job run.",
			},
			"environment_variables": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Contextual environment variables for this specific job run.",
			},
			"replicas": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of jobs to run.",
			},
		},
	}
}

func (a *StartJobDefinitionAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data StartJobDefinitionActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.jobsAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured jobsAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.JobDefinitionID.IsNull() || data.JobDefinitionID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing job_definition_id",
			"The job_definition_id attribute is required to start a job definition.",
		)

		return
	}

	jobDefinitionID := locality.ExpandID(data.JobDefinitionID.ValueString())

	var (
		region scw.Region
		err    error
	)

	if !data.Region.IsNull() && data.Region.ValueString() != "" {
		region = scw.Region(data.Region.ValueString())
	} else {
		// Try to derive region from the job_definition_id if it is a regional ID.
		if derivedRegion, id, parseErr := regional.ParseID(data.JobDefinitionID.ValueString()); parseErr == nil {
			region = derivedRegion
			jobDefinitionID = id
		} else {
			// Use default region from provider configuration
			defaultRegion, exists := a.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Missing region",
					"The region attribute is required to start a job definition. Please provide it explicitly or configure a default region in the provider.",
				)

				return
			}

			region = defaultRegion
		}
	}

	startReq := &jobs.StartJobDefinitionRequest{
		Region:          region,
		JobDefinitionID: jobDefinitionID,
	}

	if !data.Command.IsNull() && data.Command.ValueString() != "" {
		command := data.Command.ValueString()
		startReq.Command = &command
	}

	if !data.EnvironmentVariables.IsNull() {
		envVars := make(map[string]string)
		resp.Diagnostics.Append(data.EnvironmentVariables.ElementsAs(ctx, &envVars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(envVars) > 0 {
			startReq.EnvironmentVariables = &envVars
		}
	}

	if !data.Replicas.IsNull() {
		replicas := uint32(data.Replicas.ValueInt64())
		startReq.Replicas = &replicas
	}

	_, err = a.jobsAPI.StartJobDefinition(startReq, scw.WithContext(ctx))

	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Jobs StartJobDefinition action",
			fmt.Sprintf("Failed to start job definition %s: %s", jobDefinitionID, err),
		)

		return
	}
}
