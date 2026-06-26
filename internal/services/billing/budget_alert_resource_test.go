package billing_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	billingSDK "github.com/scaleway/scaleway-sdk-go/api/billing/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
)

func TestAccBudgetAlert_Basic(t *testing.T) {
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
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetAlertResourceExists(tt, "scaleway_billing_budget_alert.main"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert.main", "updated_at"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert.main", "threshold", "80"),
					resource.TestCheckResourceAttrPair("scaleway_billing_budget_alert.main", "budget_id", "scaleway_billing_budget.main", "id"),
				),
			},
			{
				ResourceName:      "scaleway_billing_budget_alert.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBudgetAlert_Update(t *testing.T) {
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
						threshold = 50
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetAlertResourceExists(tt, "scaleway_billing_budget_alert.main"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert.main", "threshold", "50"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_billing_budget" "main" {
						organization_id   = "%[1]s"
						consumption_limit = 10000
						enabled           = false
					}

					resource "scaleway_billing_budget_alert" "main" {
						budget_id = scaleway_billing_budget.main.id
						threshold = 90
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetAlertResourceExists(tt, "scaleway_billing_budget_alert.main"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert.main", "threshold", "90"),
				),
			},
		},
	})
}

func TestAccBudgetAlert_WithDefaultOrganizationID(t *testing.T) {
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
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetAlertResourceExists(tt, "scaleway_billing_budget_alert.main"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert.main", "updated_at"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert.main", "threshold", "75"),
				),
			},
		},
	})
}

func checkBudgetAlertDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_billing_budget_alert" {
				continue
			}

			billingAPI := billingSDK.NewAPI(tt.Meta.ScwClient())

			listResp, err := billingAPI.ListBudgets(&billingSDK.ListBudgetsRequest{}, scw.WithContext(context.Background()))
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return fmt.Errorf("failed to list budgets: %w", err)
			}

			deletedAlertID := rs.Primary.ID

			for _, budget := range listResp.Budgets {
				for _, alert := range budget.Alerts {
					if alert.ID == deletedAlertID {
						return fmt.Errorf("budget alert %s still exists after deletion", deletedAlertID)
					}
				}
			}
		}

		return nil
	}
}

func testAccCheckBudgetAlertResourceExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		billingAPI := billingSDK.NewAPI(tt.Meta.ScwClient())

		listResp, err := billingAPI.ListBudgets(&billingSDK.ListBudgetsRequest{}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to list budgets: %w", err)
		}

		alertID := rs.Primary.ID
		found := false

		for _, budget := range listResp.Budgets {
			for _, alert := range budget.Alerts {
				if alert.ID == alertID {
					found = true

					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return fmt.Errorf("budget alert %s not found in the list", alertID)
		}

		if rs.Primary.ID == "" {
			return errors.New("budget alert ID is not set")
		}

		return nil
	}
}
