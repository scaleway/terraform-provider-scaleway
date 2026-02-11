package cockpit

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*TriggerTestAlertAction)(nil)
	_ action.ActionWithConfigure = (*TriggerTestAlertAction)(nil)
)

type TriggerTestAlertAction struct {
	regionalAPI *cockpit.RegionalAPI
	meta        *meta.Meta
}

func (a *TriggerTestAlertAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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
	a.regionalAPI = cockpit.NewRegionalAPI(client)
	a.meta = m
}

func (a *TriggerTestAlertAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_trigger_test_alert"
}

type TriggerTestAlertActionModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	Region    types.String `tfsdk:"region"`
}

func NewTriggerTestAlertAction() action.Action {
	return &TriggerTestAlertAction{}
}

//go:embed descriptions/triggerTest_action.md
var triggerTestAlertActionDescription string

func (a *TriggerTestAlertAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: triggerTestAlertActionDescription,
		Description:         triggerTestAlertActionDescription,
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the Project",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"region": regional.SchemaAttribute("The region you want to attach the resource to"),
		},
	}
}

func (a *TriggerTestAlertAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data TriggerTestAlertActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.regionalAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured regionalAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.ProjectID.IsNull() || data.ProjectID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing project_id",
			"The project_id attribute is required to trigger a test alert.",
		)

		return
	}

	var region scw.Region

	if !data.Region.IsNull() && data.Region.ValueString() != "" {
		parsedRegion, err := scw.ParseRegion(data.Region.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid region",
				fmt.Sprintf("The region attribute must be a valid Scaleway region. Got %q: %s", data.Region.ValueString(), err),
			)

			return
		}

		region = parsedRegion
	} else {
		// Use default region from provider configuration
		defaultRegion, exists := a.meta.ScwClient().GetDefaultRegion()
		if !exists {
			resp.Diagnostics.AddError(
				"Missing region",
				"The region attribute is required to trigger a test alert. Please provide it explicitly or configure a default region in the provider.",
			)

			return
		}

		region = defaultRegion
	}

	err := a.regionalAPI.TriggerTestAlert(&cockpit.RegionalAPITriggerTestAlertRequest{
		ProjectID: data.ProjectID.ValueString(),
		Region:    region,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Cockpit TriggerTestAlert action",
			fmt.Sprintf("Failed to trigger test alert for project %s: %s", data.ProjectID.ValueString(), err),
		)

		return
	}
}
