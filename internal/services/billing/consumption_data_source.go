package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	billing "github.com/scaleway/scaleway-sdk-go/api/billing/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceConsumptions() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceBillingConsumptionsRead,
		Schema: map[string]*schema.Schema{
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
			"consumptions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Monetary value of the consumption",
						},
						"product_name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "The product name",
						},
						"category_name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Name of consumption category",
						},
						"sku": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Unique identifier of the product",
						},
						"unit": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Unit of consumed quantity",
						},
						"billed_quantity": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Consumed quantity",
						},
						"project_id": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Project ID of the consumption",
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

func DataSourceBillingConsumptionsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := billingAPI(m)

	res, err := api.ListConsumptions(&billing.ListConsumptionsRequest{
		CategoryName:   types.ExpandStringPtr(d.Get("category_name")),
		BillingPeriod:  types.ExpandStringPtr(d.Get("billing_period")),
		OrganizationID: types.ExpandStringPtr(d.Get("organization_id")),
		ProjectID:      types.ExpandStringPtr(d.Get("project_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	consumptions := []any(nil)

	for _, consumption := range res.Consumptions {
		rawConsumption := make(map[string]any)
		rawConsumption["value"] = consumption.Value.String()
		rawConsumption["product_name"] = consumption.ProductName
		rawConsumption["project_id"] = consumption.ProjectID
		rawConsumption["category_name"] = consumption.CategoryName
		rawConsumption["sku"] = consumption.Sku
		rawConsumption["unit"] = consumption.Unit
		rawConsumption["billed_quantity"] = consumption.BilledQuantity

		consumptions = append(consumptions, rawConsumption)
	}

	hashedID := sha256.Sum256([]byte(d.Get("organization_id").(string)))
	d.SetId(hex.EncodeToString(hashedID[:]))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("consumptions", consumptions)

	return nil
}
