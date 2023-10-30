package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceBillingConsumption_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_billing_consumptions" "my-consumption" {}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_billing_consumptions.my-consumption", "consumptions.0.value"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_consumptions.my-consumption", "consumptions.0.description"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_consumptions.my-consumption", "consumptions.0.project_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_consumptions.my-consumption", "consumptions.0.category"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_consumptions.my-consumption", "consumptions.0.operation_path"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_consumptions.my-consumption", "updated_at"),
				),
			},
		},
	})
}
