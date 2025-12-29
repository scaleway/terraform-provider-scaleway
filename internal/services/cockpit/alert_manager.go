package cockpit

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceCockpitAlertManager() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCockpitAlertManagerCreate,
		ReadContext:   ResourceCockpitAlertManagerRead,
		UpdateContext: ResourceCockpitAlertManagerUpdate,
		DeleteContext: ResourceCockpitAlertManagerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: alertManagerSchema,
	}
}

func alertManagerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": account.ProjectIDSchema(),
		"enable_managed_alerts": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Deprecated:  "Use 'preconfigured_alert_ids' instead. This field will be removed in a future version.",
			Description: "Enable or disable the alert manager (deprecated)",
		},
		"preconfigured_alert_ids": {
			Type:             schema.TypeSet,
			Optional:         true,
			Description:      "List of preconfigured alert rule IDs to enable explicitly. Use the scaleway_cockpit_preconfigured_alert data source to list available alerts.",
			Elem:             &schema.Schema{Type: schema.TypeString},
			DiffSuppressFunc: diffSuppressPreconfiguredAlertIDs,
		},
		"contact_points": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of contact points",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"email": {
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: verify.IsEmail(),
						Description:      "Email addresses for the alert receivers",
					},
				},
			},
		},
		"region": regional.Schema(),
		"alert_manager_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Alert manager URL",
		},
	}
}

func ResourceCockpitAlertManagerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		projectID, err = getDefaultProjectID(ctx, meta)
		if err != nil {
			return diag.FromErr(err)
		}

		_ = d.Set("project_id", projectID)
	}

	contactPoints, _ := d.Get("contact_points").([]any)
	if contactPoints == nil {
		contactPoints = []any{}
	}

	_, err = api.EnableAlertManager(&cockpit.RegionalAPIEnableAlertManagerRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		// If alert manager is already enabled, ignore the 409 error
		if !httperrors.Is409(err) {
			return diag.FromErr(err)
		}
	}

	if shouldEnableLegacyManagedAlerts(d) {
		_, err = api.EnableManagedAlerts(&cockpit.RegionalAPIEnableManagedAlertsRequest{ //nolint:staticcheck // legacy managed alerts path
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle preconfigured alerts
	if v, ok := d.GetOk("preconfigured_alert_ids"); ok {
		alertIDs := types.ExpandStrings(v.(*schema.Set).List())
		if len(alertIDs) > 0 {
			_, err = api.EnableAlertRules(&cockpit.RegionalAPIEnableAlertRulesRequest{
				Region:    region,
				ProjectID: projectID,
				RuleIDs:   alertIDs,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			// Wait for alerts to be enabled
			_, err = api.WaitForPreconfiguredAlerts(&cockpit.WaitForPreconfiguredAlertsRequest{
				Region:             region,
				ProjectID:          projectID,
				PreconfiguredRules: alertIDs,
				TargetStatus:       cockpit.AlertStatusEnabled,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if len(contactPoints) > 0 {
		for _, cp := range contactPoints {
			cpMap, ok := cp.(map[string]any)
			if !ok {
				return diag.FromErr(errors.New("invalid contact point format"))
			}

			email, ok := cpMap["email"].(string)
			if !ok {
				return diag.FromErr(errors.New("invalid email format"))
			}

			emailCP := &cockpit.ContactPointEmail{
				To: email,
			}

			_, err = api.CreateContactPoint(&cockpit.RegionalAPICreateContactPointRequest{
				ProjectID: projectID,
				Email:     emailCP,
				Region:    region,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.SetId(ResourceCockpitAlertManagerID(region, projectID))

	return ResourceCockpitAlertManagerRead(ctx, d, meta)
}

func ResourceCockpitAlertManagerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, projectID, err := NewAPIWithRegionAndProjectID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	alertManager, err := api.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) || httperrors.Is403(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	// Note: We don't set "enable_managed_alerts" here because it's automatically
	// managed by the API when preconfigured alerts are enabled/disabled.
	// Setting it would cause perpetual drift.
	_ = d.Set("region", string(alertManager.Region))
	_ = d.Set("alert_manager_url", alertManager.AlertManagerURL)
	_ = d.Set("project_id", projectID)

	var userRequestedIDs []string

	alerts, err := api.ListAlerts(&cockpit.RegionalAPIListAlertsRequest{
		Region:          region,
		ProjectID:       projectID,
		IsPreconfigured: scw.BoolPtr(true),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	alertStatusMap := make(map[string]cockpit.AlertStatus)

	for _, alert := range alerts.Alerts {
		if alert.PreconfiguredData != nil && alert.PreconfiguredData.PreconfiguredRuleID != "" {
			alertStatusMap[alert.PreconfiguredData.PreconfiguredRuleID] = alert.RuleStatus
		}
	}

	if v, ok := d.GetOk("preconfigured_alert_ids"); ok {
		requestedIDs := types.ExpandStrings(v.(*schema.Set).List())
		requestedMap := make(map[string]bool)

		for _, id := range requestedIDs {
			requestedMap[id] = true
		}

		for ruleID, status := range alertStatusMap {
			if status == cockpit.AlertStatusEnabled || status == cockpit.AlertStatusEnabling {
				if requestedMap[ruleID] {
					userRequestedIDs = append(userRequestedIDs, ruleID)
				}
			}
		}
	}

	_ = d.Set("preconfigured_alert_ids", userRequestedIDs)

	contactPoints, err := api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		// If we can't read contact points (403), treat as empty list
		if httperrors.Is403(err) {
			_ = d.Set("contact_points", []map[string]any{})
			return nil
		}
		return diag.FromErr(err)
	}

	var contactPointsList []map[string]any

	for _, cp := range contactPoints.ContactPoints {
		if cp.Email != nil {
			contactPoint := map[string]any{
				"email": cp.Email.To,
			}
			contactPointsList = append(contactPointsList, contactPoint)
		}
	}

	_ = d.Set("contact_points", contactPointsList)

	return nil
}

func ResourceCockpitAlertManagerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, projectID, err := NewAPIWithRegionAndProjectID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("preconfigured_alert_ids") {
		oldIDs, newIDs := d.GetChange("preconfigured_alert_ids")
		oldSet := oldIDs.(*schema.Set)
		newSet := newIDs.(*schema.Set)

		// IDs to disable: in old but not in new
		toDisable := types.ExpandStrings(oldSet.Difference(newSet).List())
		if len(toDisable) > 0 {
			_, err = api.DisableAlertRules(&cockpit.RegionalAPIDisableAlertRulesRequest{
				Region:    region,
				ProjectID: projectID,
				RuleIDs:   toDisable,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			// Wait for alerts to be disabled
			_, err = api.WaitForPreconfiguredAlerts(&cockpit.WaitForPreconfiguredAlertsRequest{
				Region:             region,
				ProjectID:          projectID,
				PreconfiguredRules: toDisable,
				TargetStatus:       cockpit.AlertStatusDisabled,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// IDs to enable: in new but not in old
		toEnable := types.ExpandStrings(newSet.Difference(oldSet).List())
		if len(toEnable) > 0 {
			_, err = api.EnableAlertRules(&cockpit.RegionalAPIEnableAlertRulesRequest{
				Region:    region,
				ProjectID: projectID,
				RuleIDs:   toEnable,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			// Wait for alerts to be enabled
			_, err = api.WaitForPreconfiguredAlerts(&cockpit.WaitForPreconfiguredAlertsRequest{
				Region:             region,
				ProjectID:          projectID,
				PreconfiguredRules: toEnable,
				TargetStatus:       cockpit.AlertStatusEnabled,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("enable_managed_alerts") {
		oldVal, newVal := d.GetChange("enable_managed_alerts")
		oldBool := oldVal.(bool)
		newBool := newVal.(bool)

		switch {
		case !newBool && oldBool:
			_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{ //nolint:staticcheck // legacy managed alerts path
				Region:    region,
				ProjectID: projectID,
			}, scw.WithContext(ctx))
		case newBool && shouldEnableLegacyManagedAlerts(d):
			_, err = api.EnableManagedAlerts(&cockpit.RegionalAPIEnableManagedAlertsRequest{ //nolint:staticcheck // legacy managed alerts path
				Region:    region,
				ProjectID: projectID,
			}, scw.WithContext(ctx))
		}

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("contact_points") {
		oldContactPointsInterface, newContactPointsInterface := d.GetChange("contact_points")
		oldContactPoints := oldContactPointsInterface.([]any)
		newContactPoints := newContactPointsInterface.([]any)

		oldContactMap := make(map[string]map[string]any)

		for _, oldCP := range oldContactPoints {
			cp := oldCP.(map[string]any)
			email := cp["email"].(string)
			oldContactMap[email] = cp
		}

		newContactMap := make(map[string]map[string]any)

		for _, newCP := range newContactPoints {
			cp := newCP.(map[string]any)
			email := cp["email"].(string)
			newContactMap[email] = cp
		}

		for email := range oldContactMap {
			if _, found := newContactMap[email]; !found {
				err := api.DeleteContactPoint(&cockpit.RegionalAPIDeleteContactPointRequest{
					Region:    region,
					ProjectID: projectID,
					Email:     &cockpit.ContactPointEmail{To: email},
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		for email := range newContactMap {
			if _, found := oldContactMap[email]; !found {
				contactPointEmail := &cockpit.ContactPointEmail{To: email}

				_, err = api.CreateContactPoint(&cockpit.RegionalAPICreateContactPointRequest{
					Region:    region,
					ProjectID: projectID,
					Email:     contactPointEmail,
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return ResourceCockpitAlertManagerRead(ctx, d, meta)
}

func ResourceCockpitAlertManagerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, projectID, err := NewAPIWithRegionAndProjectID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Disable all preconfigured alerts if any are enabled
	if v, ok := d.GetOk("preconfigured_alert_ids"); ok {
		alertIDs := types.ExpandStrings(v.(*schema.Set).List())
		if len(alertIDs) > 0 {
			_, err = api.DisableAlertRules(&cockpit.RegionalAPIDisableAlertRulesRequest{
				Region:    region,
				ProjectID: projectID,
				RuleIDs:   alertIDs,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.Get("enable_managed_alerts").(bool) {
		_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{ //nolint:staticcheck // legacy managed alerts path
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is403(err) && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	contactPoints, err := api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	for _, cp := range contactPoints.ContactPoints {
		if cp.Email != nil {
			err = api.DeleteContactPoint(&cockpit.RegionalAPIDeleteContactPointRequest{
				Region:    region,
				ProjectID: projectID,
				Email:     &cockpit.ContactPointEmail{To: cp.Email.To},
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	_, err = api.DisableAlertManager(&cockpit.RegionalAPIDisableAlertManagerRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

// ResourceCockpitAlertManagerID builds the resource identifier
// The resource identifier format is "Region/ProjectID/1"
func ResourceCockpitAlertManagerID(region scw.Region, projectID string) (resourceID string) {
	return fmt.Sprintf("%s/%s/1", region, projectID)
}

func shouldEnableLegacyManagedAlerts(d *schema.ResourceData) bool {
	if !d.Get("enable_managed_alerts").(bool) {
		return false
	}

	if v, ok := d.GetOk("preconfigured_alert_ids"); ok {
		if set, ok := v.(*schema.Set); ok && set.Len() > 0 {
			return false
		}
	}

	return true
}

func diffSuppressPreconfiguredAlertIDs(k, _, _ string, d *schema.ResourceData) bool {
	baseKey := strings.TrimSuffix(k, ".#")
	oldSet, newSet := d.GetChange(baseKey)

	var oldList, newList []string

	if oldSetTyped, ok := oldSet.(*schema.Set); ok {
		oldList = types.ExpandStrings(oldSetTyped.List())
	} else if oldListAny, ok := oldSet.([]any); ok {
		oldList = types.ExpandStrings(oldListAny)
	} else {
		return false
	}

	if newSetTyped, ok := newSet.(*schema.Set); ok {
		newList = types.ExpandStrings(newSetTyped.List())
	} else if newListAny, ok := newSet.([]any); ok {
		newList = types.ExpandStrings(newListAny)
	} else {
		return false
	}

	return types.CompareStringListsIgnoringOrder(oldList, newList)
}
