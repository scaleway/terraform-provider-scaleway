package edgeservices

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourcePlan() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourcePlanCreate,
		ReadContext:   ResourcePlanRead,
		UpdateContext: ResourcePlanUpdate,
		DeleteContext: ResourcePlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: verify.ValidateEnum[edgeservices.PlanName](),
				Description:      "Name of the plan",
			},
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourcePlanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	projectId, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	plan, err := api.SelectPlan(&edgeservices.SelectPlanRequest{
		ProjectID: projectId,
		PlanName:  edgeservices.PlanName(d.Get("name").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", projectId, plan.PlanName.String()))

	return nil
}

func ResourcePlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	projectId, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	plan, err := api.GetCurrentPlan(&edgeservices.GetCurrentPlanRequest{
		ProjectID: projectId,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", plan.PlanName.String())

	return nil
}

func ResourcePlanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	if d.HasChange("name") {
		projectId, _, err := meta.ExtractProjectID(d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = api.SelectPlan(&edgeservices.SelectPlanRequest{
			ProjectID: projectId,
			PlanName:  edgeservices.PlanName(d.Get("name").(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourcePlanRead(ctx, d, m)
}

func ResourcePlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	projectId, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteCurrentPlan(&edgeservices.DeleteCurrentPlanRequest{
		ProjectID: projectId,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
