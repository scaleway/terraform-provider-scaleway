package scaleway

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	billing "github.com/scaleway/scaleway-sdk-go/api/billing/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayBillingConsumptions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayBillingConsumptionsRead,
		Schema: map[string]*schema.Schema{
			"organization_id": organizationIDSchema(),
			"consumptions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"description": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"project_id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"category": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"operation_path": {
							Computed: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"updated_at": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func dataSourceScalewayBillingConsumptionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := billingAPI(meta)

	res, err := api.GetConsumption(&billing.GetConsumptionRequest{
		OrganizationID: d.Get("organization_id").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	consumptions := []interface{}(nil)
	for _, consumption := range res.Consumptions {
		rawConsumption := make(map[string]interface{})
		rawConsumption["value"] = consumption.Value.String()
		rawConsumption["description"] = consumption.Description
		rawConsumption["project_id"] = consumption.ProjectID
		rawConsumption["category"] = consumption.Category
		rawConsumption["operation_path"] = consumption.OperationPath

		consumptions = append(consumptions, rawConsumption)
	}

	hashedID := sha256.Sum256([]byte(d.Get("organization_id").(string)))
	d.SetId(fmt.Sprintf("%x", hashedID))
	_ = d.Set("updated_at", flattenTime(res.UpdatedAt))
	_ = d.Set("consumptions", consumptions)

	return nil
}
