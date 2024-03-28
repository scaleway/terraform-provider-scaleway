package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
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
	}
}

func DataSourceCockpitPlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)

	res, err := api.ListPlans(&cockpit.ListPlansRequest{}, scw.WithContext(ctx), scw.WithAllPages())
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

	d.SetId(plan.ID)
	_ = d.Set("name", plan.Name.String())

	return nil
}
