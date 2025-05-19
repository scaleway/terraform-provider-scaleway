package billing_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceConsumption_Basic(t *testing.T) {
	// somehow the project_id is not set in the query parameters locally when the organization_id is set
	tt := acctest.NewTestToolsWithoutDefaultProjectID(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_billing_consumptions" "my-consumption" {}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConsumptionsConditionalChecks("data.scaleway_billing_consumptions.my-consumption"),
				),
			},
		},
	})
}

func testAccCheckConsumptionsConditionalChecks(resourceName string) resource.TestCheckFunc {
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
