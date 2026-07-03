package billing_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	billingSDK "github.com/scaleway/scaleway-sdk-go/api/billing/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
)

func TestAccBudgetAlertNotification_Basic(t *testing.T) {
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
			checkBudgetAlertNotificationDestroyed(tt),
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

					resource "scaleway_billing_budget_alert_notification" "email" {
						budget_alert_id = scaleway_billing_budget_alert.main.id
						email_addresses = ["alerts@example.com"]
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetAlertNotificationResourceExists(tt, "scaleway_billing_budget_alert_notification.email"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert_notification.email", "id"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert_notification.email", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert_notification.email", "updated_at"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.email", "type", "email"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.email", "email_addresses.#", "1"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.email", "email_addresses.0", "alerts@example.com"),
					resource.TestCheckResourceAttrPair("scaleway_billing_budget_alert_notification.email", "budget_alert_id", "scaleway_billing_budget_alert.main", "id"),
				),
			},
			{
				ResourceName:      "scaleway_billing_budget_alert_notification.email",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					notificationID := state.RootModule().Resources["scaleway_billing_budget_alert_notification.email"].Primary.ID

					return notificationID, nil
				},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 state, got %d", len(states))
					}

					state := states[0]
					if state.Attributes["budget_alert_id"] == "" {
						return errors.New("expected budget_alert_id to be set after import")
					}

					return nil
				},
			},
		},
	})
}

func TestAccBudgetAlertNotification_SMS(t *testing.T) {
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
			checkBudgetAlertNotificationDestroyed(tt),
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

					resource "scaleway_billing_budget_alert_notification" "sms" {
						budget_alert_id   = scaleway_billing_budget_alert.main.id
						sms_phone_numbers = ["+33612345678"]
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetAlertNotificationResourceExists(tt, "scaleway_billing_budget_alert_notification.sms"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert_notification.sms", "id"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.sms", "type", "sms"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.sms", "sms_phone_numbers.#", "1"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.sms", "sms_phone_numbers.0", "+33612345678"),
				),
			},
		},
	})
}

func TestAccBudgetAlertNotification_Webhook(t *testing.T) {
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
			checkBudgetAlertNotificationDestroyed(tt),
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

					resource "scaleway_billing_budget_alert_notification" "webhook" {
						budget_alert_id = scaleway_billing_budget_alert.main.id
						webhook_urls    = ["https://example.com/webhook"]
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetAlertNotificationResourceExists(tt, "scaleway_billing_budget_alert_notification.webhook"),
					resource.TestCheckResourceAttrSet("scaleway_billing_budget_alert_notification.webhook", "id"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.webhook", "type", "webhook"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.webhook", "webhook_urls.#", "1"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.webhook", "webhook_urls.0", "https://example.com/webhook"),
				),
			},
		},
	})
}

func TestAccBudgetAlertNotification_Update(t *testing.T) {
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
			checkBudgetAlertNotificationDestroyed(tt),
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

					resource "scaleway_billing_budget_alert_notification" "email" {
						budget_alert_id = scaleway_billing_budget_alert.main.id
						email_addresses = ["alerts@example.com"]
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetAlertNotificationResourceExists(tt, "scaleway_billing_budget_alert_notification.email"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.email", "email_addresses.#", "1"),
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
						threshold = 80
					}

					resource "scaleway_billing_budget_alert_notification" "email" {
						budget_alert_id = scaleway_billing_budget_alert.main.id
						email_addresses = ["alerts@example.com", "billing@example.com"]
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBudgetAlertNotificationResourceExists(tt, "scaleway_billing_budget_alert_notification.email"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.email", "email_addresses.#", "2"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.email", "email_addresses.0", "alerts@example.com"),
					resource.TestCheckResourceAttr("scaleway_billing_budget_alert_notification.email", "email_addresses.1", "billing@example.com"),
				),
			},
		},
	})
}

func checkBudgetAlertNotificationDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_billing_budget_alert_notification" {
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

			deletedNotificationID := rs.Primary.ID

			for _, budget := range listResp.Budgets {
				for _, alert := range budget.Alerts {
					for _, notification := range alert.Notifications {
						if notification.ID == deletedNotificationID {
							return fmt.Errorf("budget alert notification %s still exists after deletion", deletedNotificationID)
						}
					}
				}
			}
		}

		return nil
	}
}

func TestAccBudgetAlertNotification_NoNotificationType(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
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

					resource "scaleway_billing_budget_alert_notification" "none" {
						budget_alert_id = scaleway_billing_budget_alert.main.id
					}
				`, orgID),
				ExpectError: regexp.MustCompile(`No attribute specified when one \(and only one\) of`),
			},
		},
	})
}

func TestAccBudgetAlertNotification_MultipleNotificationTypes(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
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

					resource "scaleway_billing_budget_alert_notification" "multiple" {
						budget_alert_id   = scaleway_billing_budget_alert.main.id
						email_addresses   = ["alerts@example.com"]
						sms_phone_numbers = ["+33612345678"]
					}
				`, orgID),
				ExpectError: regexp.MustCompile(`\d+ attributes specified when one \(and only one\) of`),
			},
		},
	})
}

func testAccCheckBudgetAlertNotificationResourceExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
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

		notificationID := rs.Primary.ID
		found := false

		for _, budget := range listResp.Budgets {
			for _, alert := range budget.Alerts {
				for _, notification := range alert.Notifications {
					if notification.ID == notificationID {
						found = true

						break
					}
				}

				if found {
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return fmt.Errorf("budget alert notification %s not found in the list", notificationID)
		}

		if rs.Primary.ID == "" {
			return errors.New("budget alert notification ID is not set")
		}

		return nil
	}
}
