package billing_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceBudget_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             checkBudgetDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_billing_budget" "main" {
						organization_id   = "%[1]s"
						consumption_limit = 10000
						enabled           = false
					}

					data "scaleway_billing_budget" "main" {
						budget_id = scaleway_billing_budget.main.id
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget.main", "budget_id", "scaleway_billing_budget.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget.main", "organization_id", "scaleway_billing_budget.main", "organization_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget.main", "consumption_limit", "scaleway_billing_budget.main", "consumption_limit"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget.main", "enabled", "scaleway_billing_budget.main", "enabled"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget.main", "created_at", "scaleway_billing_budget.main", "created_at"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget.main", "updated_at", "scaleway_billing_budget.main", "updated_at"),
				),
			},
		},
	})
}

func TestAccDataSourceBudget_WithDefaultOrganizationID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	_, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             checkBudgetDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_billing_budget" "main" {
						consumption_limit = 15000
						enabled           = false
					}

					data "scaleway_billing_budget" "main" {
						budget_id = scaleway_billing_budget.main.id
						depends_on = [scaleway_billing_budget.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget.main", "budget_id", "scaleway_billing_budget.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget.main", "consumption_limit", "scaleway_billing_budget.main", "consumption_limit"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget.main", "enabled", "scaleway_billing_budget.main", "enabled"),
				),
			},
		},
	})
}

func TestAccDataSourceBudget_InvalidBudget(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_billing_budget" "main" {
						budget_id = "00000000-0000-0000-0000-000000000000"
					}
				`,
				ExpectError: regexp.MustCompile("Could not retrieve budget"),
			},
		},
	})
}
