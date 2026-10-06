package billing_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceBudgetAlert_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			checkBudgetDestroyed(tt),
			checkBudgetAlertDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_billing_budget" "main" {
						organization_id   = "%[1]s"
						consumption_limit = 10000
						enabled           = false
					}

					resource "scaleway_billing_budget_alert" "main" {
						budget_id = scaleway_billing_budget.main.id
						threshold = 80
					}

					data "scaleway_billing_budget_alert" "main" {
						alert_id = scaleway_billing_budget_alert.main.id
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert.main", "alert_id", "scaleway_billing_budget_alert.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert.main", "budget_id", "scaleway_billing_budget_alert.main", "budget_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert.main", "threshold", "scaleway_billing_budget_alert.main", "threshold"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert.main", "created_at", "scaleway_billing_budget_alert.main", "created_at"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert.main", "updated_at", "scaleway_billing_budget_alert.main", "updated_at"),
				),
			},
		},
	})
}

func TestAccDataSourceBudgetAlert_WithDefaultOrganizationID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	_, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			checkBudgetDestroyed(tt),
			checkBudgetAlertDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_billing_budget" "main" {
						consumption_limit = 10000
						enabled           = false
					}

					resource "scaleway_billing_budget_alert" "main" {
						budget_id = scaleway_billing_budget.main.id
						threshold = 75
					}

					data "scaleway_billing_budget_alert" "main" {
						alert_id = scaleway_billing_budget_alert.main.id
						depends_on = [scaleway_billing_budget_alert.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert.main", "alert_id", "scaleway_billing_budget_alert.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert.main", "budget_id", "scaleway_billing_budget_alert.main", "budget_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert.main", "threshold", "scaleway_billing_budget_alert.main", "threshold"),
				),
			},
		},
	})
}

func TestAccDataSourceBudgetAlert_InvalidAlert(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_billing_budget_alert" "main" {
						alert_id = "00000000-0000-0000-0000-000000000000"
					}
				`,
				ExpectError: regexp.MustCompile("Budget alert.*not found"),
			},
		},
	})
}
