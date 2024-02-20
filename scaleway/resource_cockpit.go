package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayCockpit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayCockpitCreate,
		ReadContext:   resourceScalewayCockpitRead,
		UpdateContext: resourceScalewayCockpitUpdate,
		DeleteContext: resourceScalewayCockpitDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultCockpitTimeout),
			Read:    schema.DefaultTimeout(defaultCockpitTimeout),
			Delete:  schema.DefaultTimeout(defaultCockpitTimeout),
			Default: schema.DefaultTimeout(defaultCockpitTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": projectIDSchema(),
			"plan": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name or ID of the plan",
			},
			"plan_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The plan ID of the cockpit",
			},
			"endpoints": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Endpoints",
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"push_metrics_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Push url for grafana Mimir",
						},
						"push_logs_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Push url for grafana loki",
						},
					},
				},
			},
		},
	}
}

func resourceScalewayCockpitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, err := cockpitAPI(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)

	res, err := api.ActivateCockpit(&cockpit.ActivateCockpitRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if targetPlanI, ok := d.GetOk("plan"); ok {
		targetPlan := targetPlanI.(string)

		plans, err := api.ListPlans(&cockpit.ListPlansRequest{}, scw.WithContext(ctx), scw.WithAllPages())
		if err != nil {
			return diag.FromErr(err)
		}

		var planID string
		for _, plan := range plans.Plans {
			if plan.Name.String() == targetPlan || plan.ID == targetPlan {
				planID = plan.ID
				break
			}
		}

		if planID == "" {
			return diag.Errorf("plan %s not found", targetPlan)
		}

		_, err = api.SelectPlan(&cockpit.SelectPlanRequest{
			ProjectID: projectID,
			PlanID:    planID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(res.ProjectID)
	return resourceScalewayCockpitRead(ctx, d, meta)
}

func resourceScalewayCockpitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, err := cockpitAPI(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := waitForCockpit(ctx, api, d.Id(), d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("plan_id", res.Plan.ID)
	_ = d.Set("endpoints", flattenCockpitEndpoints(res.Endpoints))
	_ = d.Set("push_url", createCockpitPushURL(res.Endpoints))

	return nil
}

func resourceScalewayCockpitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, err := cockpitAPI(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Id()
	_, err = waitForCockpit(ctx, api, projectID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("plan") {
		targetPlan := cockpit.PlanNameFree.String()
		if targetPlanI, ok := d.GetOk("plan"); ok {
			targetPlan = targetPlanI.(string)
		}

		plans, err := api.ListPlans(&cockpit.ListPlansRequest{}, scw.WithContext(ctx), scw.WithAllPages())
		if err != nil {
			return diag.FromErr(err)
		}

		var planID string
		for _, plan := range plans.Plans {
			if plan.Name.String() == targetPlan || plan.ID == targetPlan {
				planID = plan.ID
				break
			}
		}

		if planID == "" {
			return diag.Errorf("plan %s not found", targetPlan)
		}

		_, err = api.SelectPlan(&cockpit.SelectPlanRequest{
			ProjectID: projectID,
			PlanID:    planID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayCockpitRead(ctx, d, meta)
}

func resourceScalewayCockpitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, err := cockpitAPI(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForCockpit(ctx, api, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_, err = api.DeactivateCockpit(&cockpit.DeactivateCockpitRequest{
		ProjectID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = waitForCockpit(ctx, api, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
