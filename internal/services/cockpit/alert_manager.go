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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
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
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "List of preconfigured alert rule IDs to enable explicitly. Use the scaleway_cockpit_preconfigured_alert data source to list available alerts.",
		Elem:        &schema.Schema{Type: schema.TypeString},
	},
	"default_preconfigured_alert_ids": {
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "List of preconfigured alert rule IDs enabled automatically by the API when alert manager is activated.",
		Elem:        &schema.Schema{Type: schema.TypeString},
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
	contactPoints := d.Get("contact_points").([]any)

	_, err = api.EnableAlertManager(&cockpit.RegionalAPIEnableAlertManagerRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// Handle preconfigured alerts
	if v, ok := d.GetOk("preconfigured_alert_ids"); ok {
		alertIDs := expandStringSet(v.(*schema.Set))
		if len(alertIDs) > 0 {
			_, err = api.EnableAlertRules(&cockpit.RegionalAPIEnableAlertRulesRequest{
				Region:    region,
				ProjectID: projectID,
				RuleIDs:   alertIDs,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			// Note: Waiting for alerts to be enabled will be handled by SDK waiters when available
			// For now, we continue without waiting as the Read function handles enabling/enabled states
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
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse the ID to get projectID
	_, projectID, err := ResourceCockpitAlertManagerParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	alertManager, err := api.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// Note: We don't set "enable_managed_alerts" here because it's automatically
	// managed by the API when preconfigured alerts are enabled/disabled.
	// Setting it would cause perpetual drift.
	_ = d.Set("region", string(alertManager.Region))
	_ = d.Set("alert_manager_url", alertManager.AlertManagerURL)
	_ = d.Set("project_id", projectID)

	// Get enabled preconfigured alerts and separate user-requested vs API-default alerts
	alerts, err := api.ListAlerts(&cockpit.RegionalAPIListAlertsRequest{
		Region:          region,
		ProjectID:       projectID,
		IsPreconfigured: scw.BoolPtr(true),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	// Build a map of alert statuses
	alertStatusMap := make(map[string]cockpit.AlertStatus)
	for _, alert := range alerts.Alerts {
		if alert.PreconfiguredData != nil && alert.PreconfiguredData.PreconfiguredRuleID != "" {
			alertStatusMap[alert.PreconfiguredData.PreconfiguredRuleID] = alert.RuleStatus
		}
	}

	// Separate user-requested alerts from API-default alerts
	var userRequestedIDs []string
	var defaultEnabledIDs []string

	if v, ok := d.GetOk("preconfigured_alert_ids"); ok {
		requestedIDs := expandStringSet(v.(*schema.Set))
		requestedMap := make(map[string]bool)
		for _, id := range requestedIDs {
			requestedMap[id] = true
		}

		// Check all enabled/enabling alerts
		for ruleID, status := range alertStatusMap {
			if status == cockpit.AlertStatusEnabled || status == cockpit.AlertStatusEnabling {
				if requestedMap[ruleID] {
					// This alert was explicitly requested by the user
					userRequestedIDs = append(userRequestedIDs, ruleID)
				} else {
					// This alert was enabled automatically by the API
					defaultEnabledIDs = append(defaultEnabledIDs, ruleID)
				}
			}
		}
	} else {
		// No alerts explicitly requested, all enabled alerts are API defaults
		for ruleID, status := range alertStatusMap {
			if status == cockpit.AlertStatusEnabled || status == cockpit.AlertStatusEnabling {
				defaultEnabledIDs = append(defaultEnabledIDs, ruleID)
			}
		}
	}

	_ = d.Set("preconfigured_alert_ids", userRequestedIDs)
	_ = d.Set("default_preconfigured_alert_ids", defaultEnabledIDs)

	contactPoints, err := api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
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
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse the ID to get projectID
	_, projectID, err := ResourceCockpitAlertManagerParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("preconfigured_alert_ids") {
		oldIDs, newIDs := d.GetChange("preconfigured_alert_ids")
		oldSet := oldIDs.(*schema.Set)
		newSet := newIDs.(*schema.Set)

		// IDs to disable: in old but not in new
		toDisable := expandStringSet(oldSet.Difference(newSet))
		if len(toDisable) > 0 {
			_, err = api.DisableAlertRules(&cockpit.RegionalAPIDisableAlertRulesRequest{
				Region:    region,
				ProjectID: projectID,
				RuleIDs:   toDisable,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			// Note: Waiting for alerts to be disabled will be handled by SDK waiters when available
		}

		// IDs to enable: in new but not in old
		toEnable := expandStringSet(newSet.Difference(oldSet))
		if len(toEnable) > 0 {
			_, err = api.EnableAlertRules(&cockpit.RegionalAPIEnableAlertRulesRequest{
				Region:    region,
				ProjectID: projectID,
				RuleIDs:   toEnable,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			// Note: Waiting for alerts to be enabled will be handled by SDK waiters when available
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
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse the ID to get projectID
	_, projectID, err := ResourceCockpitAlertManagerParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Disable all preconfigured alerts if any are enabled
	if v, ok := d.GetOk("preconfigured_alert_ids"); ok {
		alertIDs := expandStringSet(v.(*schema.Set))
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

// ResourceCockpitAlertManagerParseID extracts region and project ID from the resource identifier.
// The resource identifier format is "Region/ProjectID/1"
func ResourceCockpitAlertManagerParseID(resourceID string) (region scw.Region, projectID string, err error) {
	parts := strings.Split(resourceID, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid alert manager ID format: %s", resourceID)
	}

	return scw.Region(parts[0]), parts[1], nil
}

func expandStringSet(set *schema.Set) []string {
	result := make([]string, set.Len())
	for i, v := range set.List() {
		result[i] = v.(string)
	}

	return result
}
