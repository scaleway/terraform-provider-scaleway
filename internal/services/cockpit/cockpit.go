package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceCockpit() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCockpitCreate,
		ReadContext:   ResourceCockpitRead,
		UpdateContext: ResourceCockpitUpdate,
		DeleteContext: ResourceCockpitDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc:         cockpitSchema,
		Identity:           identity.DefaultProjectID(),
		DeprecationMessage: "The scaleway_cockpit resource is deprecated and will be removed after January 1st, 2025. Use the new specialized resources instead: scaleway_cockpit_source and scaleway_cockpit_alert_manager. For Grafana access, use the scaleway_cockpit_grafana data source with IAM authentication (the scaleway_cockpit_grafana_user resource is also deprecated).",
	}
}

func cockpitSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": account.ProjectIDSchema(),
		"plan": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "free",
			Description: "[DEPRECATED] The plan field is deprecated. Any modification or selection will have no effect.",
			Deprecated:  "The 'plan' attribute is deprecated and no longer has any effect. Future updates will remove this attribute entirely.",
		},
		"plan_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "[DEPRECATED] The plan ID of the cockpit. This field is no longer relevant.",
			Deprecated:  "The 'plan_id' attribute is deprecated and will be removed in a future release.",
		},
		"endpoints": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "[DEPRECATED] Endpoints list. Please use 'scaleway_cockpit_source' instead.",
			Deprecated:  "Use 'scaleway_cockpit_source' instead of 'endpoints'. This field will be removed in future releases.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"metrics_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The metrics URL.",
					},
					"logs_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The logs URL.",
					},
					"alertmanager_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The alertmanager URL.",
					},
					"grafana_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The grafana URL.",
					},
					"traces_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The traces URL.",
					},
				},
			},
		},
		"push_url": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "[DEPRECATED] Push_url",
			Deprecated:  "Please use `scaleway_cockpit_source` instead",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"push_metrics_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Push URL for metrics (Grafana Mimir)",
					},
					"push_logs_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Push URL for logs (Grafana Loki)",
					},
				},
			},
		},
	}
}

func ResourceCockpitCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	projectID := d.Get("project_id").(string)
	if projectID == "" {
		_, err := getDefaultProjectID(ctx, m)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err := identity.SetFlatIdentity(d, projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceCockpitRead(ctx, d, m)
}

func ResourceCockpitRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics

	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalAPI, region, err := cockpitAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		projectID, err = getDefaultProjectID(ctx, m)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Deprecated attribute: 'plan'",
		Detail:   "The 'plan' attribute is deprecated and will be removed in a future version. Any changes to this attribute will have no effect.",
	})

	_ = d.Set("plan", d.Get("plan"))
	_ = d.Set("plan_id", "")

	dataSourcesRes, err := regionalAPI.ListDataSources(&cockpit.RegionalAPIListDataSourcesRequest{
		Region:    region,
		ProjectID: projectID,
		Origin:    "custom",
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("project_id", projectID)
	err = identity.SetFlatIdentity(d, projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	grafana, err := api.GetGrafana(&cockpit.GlobalAPIGetGrafanaRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if grafana.GrafanaURL == "" {
		grafana.GrafanaURL = createGrafanaURL(projectID, region)
	}

	alertManager, err := regionalAPI.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
		ProjectID: projectID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	alertManagerURL := ""
	if alertManager.AlertManagerURL != nil {
		alertManagerURL = *alertManager.AlertManagerURL
	}

	endpoints := flattenCockpitEndpoints(dataSourcesRes.DataSources, grafana.GrafanaURL, alertManagerURL)

	_ = d.Set("endpoints", endpoints)
	_ = d.Set("push_url", createCockpitPushURLList(endpoints))

	return diags
}

func ResourceCockpitUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if d.HasChange("plan") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Deprecated attribute update: 'plan'",
			Detail:   "Updating 'plan' has no effect as it is deprecated and will be removed in a future version.",
		})
	}

	diags = append(diags, ResourceCockpitRead(ctx, d, m)...)

	return diags
}

func ResourceCockpitDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")

	return nil
}
