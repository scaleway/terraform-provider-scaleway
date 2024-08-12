package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func DataSourcePlan() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceCockpitPlanRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the plan",
				Required:    true,
			},
		},
		DeprecationMessage: "The 'Plan' data source is deprecated because it duplicates the functionality of the 'scaleway_cockpit' resource. Please use the 'scaleway_cockpit' resource instead.",
	}
}

func DataSourceCockpitPlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)

	res, err := api.ListPlans(&cockpit.GlobalAPIListPlansRequest{}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	var plan *cockpit.Plan
	for _, p := range res.Plans {
		if p.Name.String() == name {
			plan = p
			break
		}
	}

	if plan == nil {
		return diag.Errorf("could not find plan with name %s", name)
	}

	d.SetId(plan.Name.String())
	_ = d.Set("name", plan.Name.String())

	return nil
}
