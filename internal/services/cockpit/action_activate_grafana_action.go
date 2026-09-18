package cockpit

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*ActivateGrafanaAction)(nil)
	_ action.ActionWithConfigure = (*ActivateGrafanaAction)(nil)
)

type ActivateGrafanaAction struct {
	globalAPI *cockpit.GlobalAPI
	meta      *meta.Meta
}

func NewActivateGrafanaAction() action.Action {
	return &ActivateGrafanaAction{}
}

func (a *ActivateGrafanaAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

	globalAPI, err := NewGlobalAPI(m)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring Cockpit Global API",
			fmt.Sprintf("Failed to configure the Cockpit Global API: %s", err),
		)

		return
	}

	a.globalAPI = globalAPI
	a.meta = m
}

func (a *ActivateGrafanaAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_activate_grafana"
}

type ActivateGrafanaActionModel struct {
	ProjectID types.String `tfsdk:"project_id"`
}

//go:embed descriptions/activateGrafana_action.md
var activateGrafanaActionDescription string

func (a *ActivateGrafanaAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: activateGrafanaActionDescription,
		Description:         activateGrafanaActionDescription,
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the Project",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
		},
	}
}

func (a *ActivateGrafanaAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ActivateGrafanaActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.globalAPI == nil || a.meta == nil {
		resp.Diagnostics.AddError(
			"Unconfigured action",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	projectID := data.ProjectID.ValueString()
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing project_id",
			"The project_id attribute is required to activate Grafana.",
		)

		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Retrieving Grafana URL for project " + projectID,
	})

	err := activateGrafanaViaIAM(ctx, a.meta, a.globalAPI, projectID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error activating Cockpit Grafana",
			fmt.Sprintf("Failed to activate Grafana for project %s: %s", projectID, err),
		)

		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Grafana activated for project " + projectID,
	})
}

// activateGrafanaViaIAM provisions Grafana for a project by performing an IAM-authenticated
// first access against the Grafana instance. This replaces the provisioning side-effect of the
// deprecated CreateGrafanaUser API.
func activateGrafanaViaIAM(ctx context.Context, m *meta.Meta, api *cockpit.GlobalAPI, projectID string) error {
	grafana, err := api.GetGrafana(&cockpit.GlobalAPIGetGrafanaRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("get grafana: %w", err)
	}

	if grafana == nil || grafana.GrafanaURL == "" {
		return fmt.Errorf("empty grafana URL for project %s", projectID)
	}

	secretKey, hasSecretKey := m.ScwClient().GetSecretKey()
	if !hasSecretKey || secretKey == "" {
		return errors.New("missing secret key to activate grafana via IAM")
	}

	const (
		maxAttempts = 5
		retryWait   = 5 * time.Second
	)

	var lastStatus string

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, grafana.GrafanaURL+"/api/org", nil)
		if err != nil {
			return err
		}

		req.Header.Set("X-Auth-Token", secretKey)

		httpResp, err := m.HTTPClient().Do(req)
		if err != nil {
			return fmt.Errorf("access grafana: %w", err)
		}

		_, _ = io.Copy(io.Discard, httpResp.Body)
		_ = httpResp.Body.Close()

		if httpResp.StatusCode == http.StatusOK {
			return nil
		}

		lastStatus = httpResp.Status

		// Grafana may return 5xx while it is still being provisioned after first access.
		if httpResp.StatusCode < http.StatusInternalServerError || attempt == maxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryWait):
		}
	}

	return fmt.Errorf("access grafana: unexpected status %s", lastStatus)
}
