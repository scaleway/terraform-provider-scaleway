package edgeservices

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
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
		SchemaFunc:    planSchema,
		Identity:      planIdentity(),
	}
}

func planIdentity() *schema.ResourceIdentity {
	return &schema.ResourceIdentity{
		Version: 1,
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"project_id": identity.DefaultProjectIDAttribute(),
			}
		},
		IdentityUpgraders: []schema.IdentityUpgrader{
			{
				Version: 0,
				Type: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"project_id": tftypes.String,
						"name":       tftypes.String,
					},
				},
				Upgrade: upgradePlanIdentityV0,
			},
		},
	}
}

func upgradePlanIdentityV0(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	if rawState == nil {
		return map[string]any{}, nil
	}

	projectID, ok := rawState["project_id"].(string)
	if !ok {
		return nil, fmt.Errorf("identity project_id must be a string, got %T", rawState["project_id"])
	}

	return map[string]any{
		"project_id": projectID,
	}, nil
}

func planSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: verify.ValidateEnum[edgeservices.PlanName](),
			Description:      "Name of the plan",
		},
		"project_id": account.ProjectIDSchema(),
	}
}

func setPlanIdentity(d *schema.ResourceData, projectID, planName string) error {
	resourceIdentity, err := d.Identity()
	if err != nil {
		return err
	}

	if err = resourceIdentity.Set("project_id", projectID); err != nil {
		return err
	}

	d.SetId(projectID + "/" + planName)

	return nil
}

func ResourcePlanCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	projectID, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	plan, err := api.SelectPlan(&edgeservices.SelectPlanRequest{
		ProjectID: projectID,
		PlanName:  edgeservices.PlanName(d.Get("name").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("project_id", projectID)

	if err = setPlanIdentity(d, projectID, plan.PlanName.String()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourcePlanRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	idParts := identity.ParseMultiPartID(d.Id(), "project_id", "name")
	projectID := idParts["project_id"]

	// Legacy / broken state may have an empty project_id in the ID
	// Drop it so the next apply recreates with a resolved project_id
	if projectID == "" {
		d.SetId("")

		return nil
	}

	plan, err := api.GetCurrentPlan(&edgeservices.GetCurrentPlanRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if plan.PlanName == "" || plan.PlanName == edgeservices.PlanNameUnknownName {
		d.SetId("")

		return nil
	}

	_ = d.Set("name", plan.PlanName.String())
	_ = d.Set("project_id", projectID)

	if err = setPlanIdentity(d, projectID, plan.PlanName.String()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourcePlanUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	if d.HasChange("name") {
		projectID, _, err := meta.ExtractProjectID(d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = api.SelectPlan(&edgeservices.SelectPlanRequest{
			ProjectID: projectID,
			PlanName:  edgeservices.PlanName(d.Get("name").(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourcePlanRead(ctx, d, m)
}

func ResourcePlanDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	projectID, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteCurrentPlan(&edgeservices.DeleteCurrentPlanRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
