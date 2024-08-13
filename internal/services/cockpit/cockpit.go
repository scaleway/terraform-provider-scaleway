package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
		Schema: map[string]*schema.Schema{
			"project_id": account.ProjectIDSchema(),
			"plan": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name or ID of the plan",
			},
			"plan_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The plan ID of the cockpit",
				Deprecated:  "Please use Name only",
			},
			"endpoints": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Endpoints",
				Deprecated:  "Please use `scaleway_cockpit_source` instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metrics_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The metrics URL",
						},
						"logs_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The logs URL",
						},
						"alertmanager_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The alertmanager URL",
						},
						"grafana_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The grafana URL",
						},
						"traces_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The traces URL",
						},
					},
				},
			},
			"push_url": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Push_url",
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
		},
	}
}

func ResourceCockpitCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if targetPlanI, ok := d.GetOk("plan"); ok {
		targetPlan := targetPlanI.(string)

		plans, err := api.ListPlans(&cockpit.GlobalAPIListPlansRequest{}, scw.WithContext(ctx), scw.WithAllPages())
		if err != nil {
			return diag.FromErr(err)
		}

		var planName string
		for _, plan := range plans.Plans {
			if plan.Name.String() == targetPlan {
				planName = plan.Name.String()
				break
			}
		}

		if planName == "" {
			return diag.Errorf("plan %s not found", targetPlan)
		}

		_, err = api.SelectPlan(&cockpit.GlobalAPISelectPlanRequest{
			ProjectID: projectID,
			PlanName:  cockpit.PlanName(planName),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(projectID)
	return ResourceCockpitRead(ctx, d, m)
}

func ResourceCockpitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalAPI, region, err := cockpitAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.GetCurrentPlan(&cockpit.GlobalAPIGetCurrentPlanRequest{
		ProjectID: d.Get("project_id").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("project_id", d.Get("project_id").(string))
	_ = d.Set("plan", res.Name.String())
	_ = d.Set("plan_id", res.Name.String())

	dataSourcesRes, err := regionalAPI.ListDataSources(&cockpit.RegionalAPIListDataSourcesRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Origin:    "external",
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	grafana, err := api.GetGrafana(&cockpit.GlobalAPIGetGrafanaRequest{
		ProjectID: d.Get("project_id").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	alertManager, err := regionalAPI.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
		ProjectID: d.Get("project_id").(string),
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

	return nil
}

func ResourceCockpitUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Id()

	if d.HasChange("plan") {
		targetPlan := cockpit.PlanNameFree.String()
		if targetPlanI, ok := d.GetOk("plan"); ok {
			targetPlan = targetPlanI.(string)
		}

		plans, err := api.ListPlans(&cockpit.GlobalAPIListPlansRequest{}, scw.WithContext(ctx), scw.WithAllPages())
		if err != nil {
			return diag.FromErr(err)
		}

		var planName string
		for _, plan := range plans.Plans {
			if plan.Name.String() == targetPlan {
				planName = plan.Name.String()
				break
			}
		}

		if planName == "" {
			return diag.Errorf("plan %s not found", targetPlan)
		}

		_, err = api.SelectPlan(&cockpit.GlobalAPISelectPlanRequest{
			ProjectID: projectID,
			PlanName:  cockpit.PlanName(planName),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceCockpitRead(ctx, d, m)
}

func ResourceCockpitDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
