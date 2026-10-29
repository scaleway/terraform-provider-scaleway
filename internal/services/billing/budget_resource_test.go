package billing_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	billingSDK "github.com/scaleway/scaleway-sdk-go/api/billing/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
)

func TestAccBudget_Basic(t *testing.T) {
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
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetResourceExists(tt, "scaleway_billing_budget.main"),
					resource.TestCheckResourceAttr("scaleway_billing_budget.main", "organization_id", orgID),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget.main", "updated_at"),
					resource.TestCheckResourceAttr("scaleway_billing_budget.main", "consumption_limit", "10000"),
					resource.TestCheckResourceAttr("scaleway_billing_budget.main", "enabled", "false"),
				),
			},
			{
				ResourceName:      "scaleway_billing_budget.main",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					budgetID := state.RootModule().Resources["scaleway_billing_budget.main"].Primary.ID

					return budgetID, nil
				},
			},
		},
	})
}

func TestAccBudget_Update(t *testing.T) {
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
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetResourceExists(tt, "scaleway_billing_budget.main"),
					resource.TestCheckResourceAttr("scaleway_billing_budget.main", "consumption_limit", "10000"),
					resource.TestCheckResourceAttr("scaleway_billing_budget.main", "enabled", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_billing_budget" "main" {
						organization_id   = "%[1]s"
						consumption_limit = 20000
						enabled           = false
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetResourceExists(tt, "scaleway_billing_budget.main"),
					resource.TestCheckResourceAttr("scaleway_billing_budget.main", "consumption_limit", "20000"),
					resource.TestCheckResourceAttr("scaleway_billing_budget.main", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccBudget_WithDefaultOrganizationID(t *testing.T) {
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
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetResourceExists(tt, "scaleway_billing_budget.main"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget.main", "updated_at"),
					resource.TestCheckResourceAttr("scaleway_billing_budget.main", "consumption_limit", "15000"),
					resource.TestCheckResourceAttr("scaleway_billing_budget.main", "enabled", "false"),
				),
			},
		},
	})
}

func checkBudgetDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_billing_budget" {
				continue
			}

			billingAPI := billingSDK.NewAPI(tt.Meta.ScwClient())

			_, err := (&retry.StateChangeConf{
				Pending: []string{"exists"},
				Target:  []string{"deleted"},
				Refresh: func() (any, string, error) {
					_, err := billingAPI.GetBudget(&billingSDK.GetBudgetRequest{
						BudgetID: rs.Primary.ID,
					}, scw.WithContext(context.Background()))
					if err != nil {
						if httperrors.Is404(err) {
							return nil, "deleted", nil
						}

						return nil, "", err
					}

					return nil, "exists", nil
				},
				Timeout:    10 * time.Second,
				Delay:      0,
				MinTimeout: 2 * time.Second,
			}).WaitForStateContext(context.Background())
			if err == nil {
				return fmt.Errorf("budget %s still exists after deletion", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckBudgetResourceExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		billingAPI := billingSDK.NewAPI(tt.Meta.ScwClient())

		_, err := billingAPI.GetBudget(&billingSDK.GetBudgetRequest{
			BudgetID: rs.Primary.ID,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to get budget: %w", err)
		}

		if rs.Primary.ID == "" {
			return errors.New("budget ID is not set")
		}

		return nil
	}
}
