package cockpit

import (
	"context"
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/cockpit_alert_manager_resource.md
var cockpitAlertManagerResourceDescription string

var (
	_ resource.Resource                = (*AlertManagerResource)(nil)
	_ resource.ResourceWithConfigure   = (*AlertManagerResource)(nil)
	_ resource.ResourceWithImportState = (*AlertManagerResource)(nil)
	_ resource.ResourceWithIdentity    = (*AlertManagerResource)(nil)
)

func NewAlertManagerResource() resource.Resource {
	return &AlertManagerResource{}
}

type AlertManagerResource struct {
	api  *cockpit.RegionalAPI
	meta *meta.Meta
}

type contactPointModel struct {
	Email types.String `tfsdk:"email"`
}

type alertManagerResourceModel struct {
	PreconfiguredAlertIDs types.Set           `tfsdk:"preconfigured_alert_ids"`
	ID                    types.String        `tfsdk:"id"`
	ProjectID             types.String        `tfsdk:"project_id"`
	Region                types.String        `tfsdk:"region"`
	AlertManagerURL       types.String        `tfsdk:"alert_manager_url"`
	SRN                   types.String        `tfsdk:"srn"`
	ContactPoints         []contactPointModel `tfsdk:"contact_points"`
	EnableManagedAlerts   types.Bool          `tfsdk:"enable_managed_alerts"`
}

type alertManagerIdentityModel struct {
	Region    types.String `tfsdk:"region"`
	ProjectID types.String `tfsdk:"project_id"`
}

func (r *AlertManagerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_alert_manager"
}

func (r *AlertManagerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitAlertManagerResourceDescription,
		Description:         cockpitAlertManagerResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": account.ProjectIDSchemaFw("The project_id you want to attach the resource to"),
			"enable_managed_alerts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Use 'preconfigured_alert_ids' instead. This field will be removed in a future version.",
				MarkdownDescription: "Enable or disable the alert manager (deprecated)",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"preconfigured_alert_ids": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of preconfigured alert rule IDs to enable explicitly. Use the scaleway_cockpit_preconfigured_alert data source to list available alerts.",
			},
			"region": regional.SchemaAttributeComputed(),
			"alert_manager_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Alert manager URL",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"srn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Scaleway Resource Name (SRN) of the alert manager",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"contact_points": schema.ListNestedBlock{
				MarkdownDescription: "A list of contact points",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Email addresses for the alert receivers",
							Validators:          []validator.String{verify.IsStringEmail()},
						},
					},
				},
			},
		},
	}
}

func (r *AlertManagerResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.CompositeRegional("project_id")
}

func (r *AlertManagerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = cockpit.NewRegionalAPI(m.ScwClient())
}

func (r *AlertManagerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertManagerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(plan.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	alertIDs, diags := expandAlertIDs(ctx, plan.PreconfiguredAlertIDs)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err = transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.AlertManager, error) {
		return r.api.EnableAlertManager(&cockpit.RegionalAPIEnableAlertManagerRequest{
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	})
	// If alert manager is already enabled, ignore the 409 error
	if err != nil && !httperrors.Is409(err) {
		resp.Diagnostics.AddError("Failed to enable alert manager", err.Error())

		return
	}

	if shouldEnableLegacyManagedAlerts(plan.EnableManagedAlerts, alertIDs) {
		err = r.enableManagedAlerts(ctx, region, projectID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to enable managed alerts", err.Error())

			return
		}
	}

	err = r.enableAlertRules(ctx, region, projectID, alertIDs)
	if err != nil {
		resp.Diagnostics.AddError("Failed to enable preconfigured alerts", err.Error())

		return
	}

	for _, cp := range plan.ContactPoints {
		err = r.createContactPoint(ctx, region, projectID, cp.Email.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to create contact point", err.Error())

			return
		}
	}

	// enable_managed_alerts is Optional+Computed and never read back from the API.
	if plan.EnableManagedAlerts.IsUnknown() {
		plan.EnableManagedAlerts = types.BoolValue(false)
	}

	plan.ID = types.StringValue(alertManagerID(region, projectID))
	plan.Region = types.StringValue(region.String())
	plan.ProjectID = types.StringValue(projectID)

	resp.Diagnostics.Append(resp.Identity.Set(ctx, alertManagerIdentity(region, projectID))...)

	r.refresh(ctx, region, projectID, plan, &resp.State, &resp.Diagnostics)
}

func (r *AlertManagerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    alertManagerResourceModel
		identity alertManagerIdentityModel
	)

	identityAvailable := !req.Identity.Get(ctx, &identity).HasError() && !identity.ProjectID.IsNull() && !identity.ProjectID.IsUnknown()

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, projectID, err := resolveAlertManagerID(identity, identityAvailable, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse alert manager ID", err.Error())

		return
	}

	// enable_managed_alerts is never read back from the API: default it so an imported resource does not drift.
	if state.EnableManagedAlerts.IsNull() {
		state.EnableManagedAlerts = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, alertManagerIdentity(region, projectID))...)

	r.refresh(ctx, region, projectID, state, &resp.State, &resp.Diagnostics)
}

func (r *AlertManagerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state alertManagerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, projectID, err := parseAlertManagerID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse alert manager ID", err.Error())

		return
	}

	oldIDs, diags := expandAlertIDs(ctx, state.PreconfiguredAlertIDs)
	resp.Diagnostics.Append(diags...)

	newIDs, diags := expandAlertIDs(ctx, plan.PreconfiguredAlertIDs)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	toDisable := slices.DeleteFunc(slices.Clone(oldIDs), func(id string) bool { return slices.Contains(newIDs, id) })
	toEnable := slices.DeleteFunc(slices.Clone(newIDs), func(id string) bool { return slices.Contains(oldIDs, id) })

	err = r.disableAlertRules(ctx, region, projectID, toDisable, true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to disable preconfigured alerts", err.Error())

		return
	}

	err = r.enableAlertRules(ctx, region, projectID, toEnable)
	if err != nil {
		resp.Diagnostics.AddError("Failed to enable preconfigured alerts", err.Error())

		return
	}

	err = r.updateManagedAlerts(ctx, region, projectID, state.EnableManagedAlerts, plan.EnableManagedAlerts, newIDs)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update managed alerts", err.Error())

		return
	}

	err = r.updateContactPoints(ctx, region, projectID, state.ContactPoints, plan.ContactPoints)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update contact points", err.Error())

		return
	}

	if plan.EnableManagedAlerts.IsUnknown() {
		plan.EnableManagedAlerts = types.BoolValue(false)
	}

	r.refresh(ctx, region, projectID, plan, &resp.State, &resp.Diagnostics)
}

func (r *AlertManagerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var (
		state    alertManagerResourceModel
		identity alertManagerIdentityModel
	)

	identityAvailable := !req.Identity.Get(ctx, &identity).HasError() && !identity.ProjectID.IsNull() && !identity.ProjectID.IsUnknown()

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, projectID, err := resolveAlertManagerID(identity, identityAvailable, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse alert manager ID", err.Error())

		return
	}

	alertIDs, diags := expandAlertIDs(ctx, state.PreconfiguredAlertIDs)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Disable all preconfigured alerts if any are enabled
	err = r.disableAlertRules(ctx, region, projectID, alertIDs, false)
	if err != nil {
		resp.Diagnostics.AddError("Failed to disable preconfigured alerts", err.Error())

		return
	}

	if state.EnableManagedAlerts.ValueBool() {
		_, err = transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.AlertManager, error) {
			return r.api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{ //nolint:staticcheck // legacy managed alerts path
				Region:    region,
				ProjectID: projectID,
			}, scw.WithContext(ctx))
		})
		if err != nil && !httperrors.Is403(err) && !httperrors.Is404(err) {
			resp.Diagnostics.AddError("Failed to disable managed alerts", err.Error())

			return
		}
	}

	contactPoints, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.ListContactPointsResponse, error) {
		return r.api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list contact points", err.Error())

		return
	}

	for _, cp := range contactPoints.ContactPoints {
		if cp.Email == nil {
			continue
		}

		err = r.deleteContactPoint(ctx, region, projectID, cp.Email.To)
		if err != nil {
			resp.Diagnostics.AddError("Failed to delete contact point", err.Error())

			return
		}
	}

	_, err = transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.AlertManager, error) {
		return r.api.DisableAlertManager(&cockpit.RegionalAPIDisableAlertManagerRequest{
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	})
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to disable alert manager", err.Error())
	}
}

func (r *AlertManagerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	region, projectID, err := parseAlertManagerID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region.String())...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, alertManagerIdentity(region, projectID))...)
}

// refresh reads the alert manager and stores it in state. A missing alert manager removes the resource from state.
// prior carries the values that the API does not return (enable_managed_alerts) and the requested alert IDs.
func (r *AlertManagerResource) refresh(ctx context.Context, region scw.Region, projectID string, prior alertManagerResourceModel, state *tfsdk.State, diags *diag.Diagnostics) {
	alertManager, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.AlertManager, error) {
		return r.api.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		if httperrors.Is404(err) {
			state.RemoveResource(ctx)

			return
		}

		diags.AddError("Failed to read alert manager", err.Error())

		return
	}

	alerts, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.ListAlertsResponse, error) {
		return r.api.ListAlerts(&cockpit.RegionalAPIListAlertsRequest{
			Region:          region,
			ProjectID:       projectID,
			IsPreconfigured: new(true),
		}, scw.WithContext(ctx), scw.WithAllPages())
	})
	if err != nil {
		diags.AddError("Failed to list preconfigured alerts", err.Error())

		return
	}

	requestedIDs, d := expandAlertIDs(ctx, prior.PreconfiguredAlertIDs)
	diags.Append(d...)

	enabledIDs := []string{}

	for _, alert := range alerts.Alerts {
		if alert.PreconfiguredData == nil || !slices.Contains(requestedIDs, alert.PreconfiguredData.PreconfiguredRuleID) {
			continue
		}

		if alert.RuleStatus == cockpit.AlertStatusEnabled || alert.RuleStatus == cockpit.AlertStatusEnabling {
			enabledIDs = append(enabledIDs, alert.PreconfiguredData.PreconfiguredRuleID)
		}
	}

	alertIDs := types.SetNull(types.StringType)
	if len(enabledIDs) > 0 || !prior.PreconfiguredAlertIDs.IsNull() {
		alertIDs, d = types.SetValueFrom(ctx, types.StringType, enabledIDs)
		diags.Append(d...)
	}

	contactPoints, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.ListContactPointsResponse, error) {
		return r.api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		diags.AddError("Failed to list contact points", err.Error())

		return
	}

	model := alertManagerResourceModel{
		ContactPoints:         []contactPointModel{},
		PreconfiguredAlertIDs: alertIDs,
		ID:                    types.StringValue(alertManagerID(region, projectID)),
		ProjectID:             types.StringValue(projectID),
		Region:                types.StringValue(alertManager.Region.String()),
		AlertManagerURL:       types.StringPointerValue(alertManager.AlertManagerURL),
		SRN:                   types.StringValue(alertManager.Srn),
		EnableManagedAlerts:   prior.EnableManagedAlerts,
	}

	for _, cp := range contactPoints.ContactPoints {
		if cp.Email != nil {
			model.ContactPoints = append(model.ContactPoints, contactPointModel{Email: types.StringValue(cp.Email.To)})
		}
	}

	if diags.HasError() {
		return
	}

	diags.Append(state.Set(ctx, &model)...)
}

// updateManagedAlerts toggles the legacy managed alerts when enable_managed_alerts changed.
func (r *AlertManagerResource) updateManagedAlerts(ctx context.Context, region scw.Region, projectID string, oldValue, newValue types.Bool, newAlertIDs []string) error {
	if oldValue.Equal(newValue) {
		return nil
	}

	switch {
	case !newValue.ValueBool() && oldValue.ValueBool():
		_, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.AlertManager, error) {
			return r.api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{ //nolint:staticcheck // legacy managed alerts path
				Region:    region,
				ProjectID: projectID,
			}, scw.WithContext(ctx))
		})

		return err
	case newValue.ValueBool() && shouldEnableLegacyManagedAlerts(newValue, newAlertIDs):
		return r.enableManagedAlerts(ctx, region, projectID)
	}

	return nil
}

// updateContactPoints deletes the contact points removed from the configuration and creates the new ones.
func (r *AlertManagerResource) updateContactPoints(ctx context.Context, region scw.Region, projectID string, oldPoints, newPoints []contactPointModel) error {
	emails := func(points []contactPointModel) []string {
		result := make([]string, 0, len(points))
		for _, cp := range points {
			result = append(result, cp.Email.ValueString())
		}

		return result
	}

	oldEmails, newEmails := emails(oldPoints), emails(newPoints)

	for _, email := range oldEmails {
		if !slices.Contains(newEmails, email) {
			err := r.deleteContactPoint(ctx, region, projectID, email)
			if err != nil {
				return err
			}
		}
	}

	for _, email := range newEmails {
		if !slices.Contains(oldEmails, email) {
			err := r.createContactPoint(ctx, region, projectID, email)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *AlertManagerResource) enableManagedAlerts(ctx context.Context, region scw.Region, projectID string) error {
	_, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.AlertManager, error) {
		return r.api.EnableManagedAlerts(&cockpit.RegionalAPIEnableManagedAlertsRequest{ //nolint:staticcheck // legacy managed alerts path
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	})

	return err
}

// enableAlertRules enables the given preconfigured alert rules and waits until they are enabled.
func (r *AlertManagerResource) enableAlertRules(ctx context.Context, region scw.Region, projectID string, ruleIDs []string) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	_, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.EnableAlertRulesResponse, error) {
		return r.api.EnableAlertRules(&cockpit.RegionalAPIEnableAlertRulesRequest{
			Region:    region,
			ProjectID: projectID,
			RuleIDs:   ruleIDs,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		return err
	}

	return r.waitForPreconfiguredAlerts(ctx, region, projectID, ruleIDs, cockpit.AlertStatusEnabled)
}

// disableAlertRules disables the given preconfigured alert rules, optionally waiting until they are disabled.
func (r *AlertManagerResource) disableAlertRules(ctx context.Context, region scw.Region, projectID string, ruleIDs []string, wait bool) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	_, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.DisableAlertRulesResponse, error) {
		return r.api.DisableAlertRules(&cockpit.RegionalAPIDisableAlertRulesRequest{
			Region:    region,
			ProjectID: projectID,
			RuleIDs:   ruleIDs,
		}, scw.WithContext(ctx))
	})
	if err != nil || !wait {
		return err
	}

	return r.waitForPreconfiguredAlerts(ctx, region, projectID, ruleIDs, cockpit.AlertStatusDisabled)
}

func (r *AlertManagerResource) waitForPreconfiguredAlerts(ctx context.Context, region scw.Region, projectID string, ruleIDs []string, target cockpit.AlertStatus) error {
	_, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) ([]*cockpit.Alert, error) {
		return r.api.WaitForPreconfiguredAlerts(&cockpit.WaitForPreconfiguredAlertsRequest{
			Region:             region,
			ProjectID:          projectID,
			PreconfiguredRules: ruleIDs,
			TargetStatus:       target,
		}, scw.WithContext(ctx))
	})

	return err
}

func (r *AlertManagerResource) createContactPoint(ctx context.Context, region scw.Region, projectID, email string) error {
	_, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.ContactPoint, error) {
		return r.api.CreateContactPoint(&cockpit.RegionalAPICreateContactPointRequest{
			Region:    region,
			ProjectID: projectID,
			Email:     &cockpit.ContactPointEmail{To: email},
		}, scw.WithContext(ctx))
	})

	return err
}

func (r *AlertManagerResource) deleteContactPoint(ctx context.Context, region scw.Region, projectID, email string) error {
	return transport.RetryOnStatus(ctx, transport.IAMPropagation, func(ctx context.Context) error {
		return r.api.DeleteContactPoint(&cockpit.RegionalAPIDeleteContactPointRequest{
			Region:    region,
			ProjectID: projectID,
			Email:     &cockpit.ContactPointEmail{To: email},
		}, scw.WithContext(ctx))
	})
}

func expandAlertIDs(ctx context.Context, ids types.Set) ([]string, diag.Diagnostics) {
	var result []string

	if ids.IsNull() || ids.IsUnknown() {
		return result, nil
	}

	diags := ids.ElementsAs(ctx, &result, false)

	return result, diags
}

func alertManagerIdentity(region scw.Region, projectID string) alertManagerIdentityModel {
	return alertManagerIdentityModel{
		Region:    types.StringValue(region.String()),
		ProjectID: types.StringValue(projectID),
	}
}

// resolveAlertManagerID resolves the region and project ID from the identity when available, else from the state ID.
func resolveAlertManagerID(identity alertManagerIdentityModel, identityAvailable bool, stateID types.String) (scw.Region, string, error) {
	if identityAvailable {
		return scw.Region(identity.Region.ValueString()), identity.ProjectID.ValueString(), nil
	}

	return parseAlertManagerID(stateID.ValueString())
}

// parseAlertManagerID splits the "region/projectID/1" alert manager ID.
func parseAlertManagerID(id string) (scw.Region, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid alert manager ID format: %s, expected region/projectID/1", id)
	}

	return scw.Region(parts[0]), parts[1], nil
}

// alertManagerID builds the resource identifier, in the "region/projectID/1" format.
func alertManagerID(region scw.Region, projectID string) string {
	return fmt.Sprintf("%s/%s/1", region, projectID)
}

// shouldEnableLegacyManagedAlerts tells whether the deprecated managed alerts must be enabled:
// only when requested and no preconfigured alert is explicitly enabled.
func shouldEnableLegacyManagedAlerts(enableManagedAlerts types.Bool, alertIDs []string) bool {
	return enableManagedAlerts.ValueBool() && len(alertIDs) == 0
}
