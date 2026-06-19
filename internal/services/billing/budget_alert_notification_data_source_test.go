package billing_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceBudgetAlertNotification_Basic(t *testing.T) {
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

					data "scaleway_billing_budget_alert_notification" "main" {
						notification_id = scaleway_billing_budget_alert_notification.email.id
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert_notification.main", "notification_id", "scaleway_billing_budget_alert_notification.email", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert_notification.main", "budget_alert_id", "scaleway_billing_budget_alert_notification.email", "budget_alert_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert_notification.main", "type", "scaleway_billing_budget_alert_notification.email", "type"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert_notification.main", "created_at", "scaleway_billing_budget_alert_notification.email", "created_at"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert_notification.main", "updated_at", "scaleway_billing_budget_alert_notification.email", "updated_at"),
				),
			},
		},
	})
}

func TestAccDataSourceBudgetAlertNotification_SMS(t *testing.T) {
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

					data "scaleway_billing_budget_alert_notification" "sms" {
						notification_id = scaleway_billing_budget_alert_notification.sms.id
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert_notification.sms", "notification_id", "scaleway_billing_budget_alert_notification.sms", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_billing_budget_alert_notification.sms", "type", "scaleway_billing_budget_alert_notification.sms", "type"),
				),
			},
		},
	})
}

func TestAccDataSourceBudgetAlertNotification_InvalidNotification(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_billing_budget_alert_notification" "main" {
						notification_id = "00000000-0000-0000-0000-000000000000"
					}
				`,
				ExpectError: regexp.MustCompile("Budget alert notification.*not found"),
			},
		},
	})
}
