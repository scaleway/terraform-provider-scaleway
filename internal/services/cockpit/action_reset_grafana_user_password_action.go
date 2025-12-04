package cockpit

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ action.Action              = (*ResetGrafanaUserPasswordAction)(nil)
	_ action.ActionWithConfigure = (*ResetGrafanaUserPasswordAction)(nil)
)

type ResetGrafanaUserPasswordAction struct {
	globalAPI *cockpit.GlobalAPI
}

func (a *ResetGrafanaUserPasswordAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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
	a.globalAPI = cockpit.NewGlobalAPI(client)
}

func (a *ResetGrafanaUserPasswordAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_reset_grafana_user_password_action"
}

type ResetGrafanaUserPasswordActionModel struct {
	GrafanaUserID types.String `tfsdk:"grafana_user_id"`
	ProjectID     types.String `tfsdk:"project_id"`
}

func NewResetGrafanaUserPasswordAction() action.Action {
	return &ResetGrafanaUserPasswordAction{}
}

func (a *ResetGrafanaUserPasswordAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"grafana_user_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the Grafana user",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the Project. If not provided, will be extracted from grafana_user_id if it's in format 'project_id/user_id'",
			},
		},
	}
}

func (a *ResetGrafanaUserPasswordAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ResetGrafanaUserPasswordActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.globalAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured globalAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	grafanaUserIDStr := data.GrafanaUserID.ValueString()
	if grafanaUserIDStr == "" {
		resp.Diagnostics.AddError(
			"Missing grafana_user_id",
			"The grafana_user_id attribute is required to reset Grafana user password.",
		)

		return
	}

	// Parse ID format: project_id/grafana_user_id or just grafana_user_id
	var grafanaUserID uint32
	projectID := data.ProjectID.ValueString()

	if strings.Contains(grafanaUserIDStr, "/") {
		// ID format: project_id/grafana_user_id
		parsedProjectID, grafanaUserIDPart, err := parseCockpitID(grafanaUserIDStr)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid grafana_user_id format",
				fmt.Sprintf("The grafana_user_id must be in format 'project_id/user_id' or just 'user_id': %s", err),
			)

			return
		}

		// Use parsed project_id if project_id was not explicitly provided
		if projectID == "" {
			projectID = parsedProjectID
		}

		grafanaUserIDUint, err := strconv.ParseUint(grafanaUserIDPart, 10, 32)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid grafana_user_id",
				fmt.Sprintf("The grafana_user_id must be a valid uint32: %s", err),
			)

			return
		}

		grafanaUserID = uint32(grafanaUserIDUint)
	} else {
		// Just grafana_user_id (uint32 as string)
		if projectID == "" {
			resp.Diagnostics.AddError(
				"Missing project_id",
				"The project_id attribute is required when grafana_user_id is not in format 'project_id/user_id'.",
			)

			return
		}

		grafanaUserIDUint, err := strconv.ParseUint(grafanaUserIDStr, 10, 32)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid grafana_user_id",
				fmt.Sprintf("The grafana_user_id must be a valid uint32: %s", err),
			)

			return
		}

		grafanaUserID = uint32(grafanaUserIDUint)
	}

	_, err := a.globalAPI.ResetGrafanaUserPassword(&cockpit.GlobalAPIResetGrafanaUserPasswordRequest{
		GrafanaUserID: grafanaUserID,
		ProjectID:     projectID,
	}, scw.WithContext(ctx))

	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Cockpit ResetGrafanaUserPassword action",
			fmt.Sprintf("Failed to reset password for Grafana user %s: %s", data.GrafanaUserID.ValueString(), err),
		)

		return
	}
}

