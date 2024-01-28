package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
					testAccCheckScalewayConsumptionsConditionalChecks("data.scaleway_billing_consumptions.my-consumption"),
				),
			},
		},
	})
}

func testAccCheckScalewayConsumptionsConditionalChecks(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		attr, ok := rs.Primary.Attributes["consumptions.#"]
		if ok && attr != "0" {
			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttrSet(resourceName, "consumptions.0.value"),
				resource.TestCheckResourceAttrSet(resourceName, "consumptions.0.product_name"),
				resource.TestCheckResourceAttrSet(resourceName, "consumptions.0.project_id"),
				resource.TestCheckResourceAttrSet(resourceName, "consumptions.0.category_name"),
				resource.TestCheckResourceAttrSet(resourceName, "consumptions.0.sku"),
				resource.TestCheckResourceAttrSet(resourceName, "consumptions.0.unit"),
				resource.TestCheckResourceAttrSet(resourceName, "consumptions.0.billed_quantity"),
				resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
			}

			for _, check := range checks {
				if err := check(s); err != nil {
					return err
				}
			}
		}

		return nil
	}
}
