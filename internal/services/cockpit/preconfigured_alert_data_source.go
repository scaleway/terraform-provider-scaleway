package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCockpitPreconfiguredAlert() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCockpitPreconfiguredAlertRead,
		Schema: map[string]*schema.Schema{
			"project_id": account.ProjectIDSchema(),
			"region":     regional.Schema(),
			"alerts": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of preconfigured alerts",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the alert",
						},
						"rule": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "PromQL rule defining the alert condition",
						},
						"duration": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Duration for which the alert must be active before firing",
						},
						"rule_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the alert (enabled, disabled, enabling, disabling)",
						},
						"state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current state of the alert (inactive, pending, firing)",
						},
						"annotations": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Annotations for the alert",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"preconfigured_rule_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the preconfigured rule",
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Human readable name of the alert",
						},
						"display_description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Human readable description of the alert",
						},
						"product_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Product associated with the alert",
						},
						"product_family": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Family of the product associated with the alert",
						},
						"data_source_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the data source containing the alert rule",
						},
					},
				},
			},
			"data_source_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Filter alerts by data source ID",
				ValidateDiagFunc: verify.IsUUID(),
			},
			"rule_status": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Filter alerts by rule status (enabled, disabled)",
				ValidateDiagFunc: verify.ValidateEnum[cockpit.AlertStatus](),
			},
		},
	}
}

func dataSourceCockpitPreconfiguredAlertRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		defaultProjectID, err := getDefaultProjectID(ctx, m)
		if err != nil {
			return diag.FromErr(err)
		}

		projectID = defaultProjectID
	}

	req := &cockpit.RegionalAPIListAlertsRequest{
		Region:          region,
		ProjectID:       projectID,
		IsPreconfigured: new(true),
	}

	if dataSourceID, ok := d.GetOk("data_source_id"); ok {
		req.DataSourceID = new(dataSourceID.(string))
	}

	if ruleStatus, ok := d.GetOk("rule_status"); ok {
		status := cockpit.AlertStatus(ruleStatus.(string))
		req.RuleStatus = &status
	}

	response, err := api.ListAlerts(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	alerts := make([]map[string]any, 0, len(response.Alerts))
	for _, alert := range response.Alerts {
		alertMap := map[string]any{
			"name":           alert.Name,
			"rule":           alert.Rule,
			"duration":       alert.Duration,
			"rule_status":    string(alert.RuleStatus),
			"annotations":    alert.Annotations,
			"data_source_id": alert.DataSourceID,
		}

		if alert.State != nil {
			alertMap["state"] = string(*alert.State)
		}

		if alert.PreconfiguredData != nil {
			alertMap["preconfigured_rule_id"] = alert.PreconfiguredData.PreconfiguredRuleID
			alertMap["display_name"] = alert.PreconfiguredData.DisplayName
			alertMap["display_description"] = alert.PreconfiguredData.DisplayDescription
			alertMap["product_name"] = alert.PreconfiguredData.ProductName
			alertMap["product_family"] = alert.PreconfiguredData.ProductFamily
		}

		alerts = append(alerts, alertMap)
	}

	d.SetId(regional.NewIDString(region, projectID))
	_ = d.Set("project_id", projectID)
	_ = d.Set("region", string(region))
	_ = d.Set("alerts", alerts)

	return nil
}
