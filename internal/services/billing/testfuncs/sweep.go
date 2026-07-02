package billingtestfuncs

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	billingSDK "github.com/scaleway/scaleway-sdk-go/api/billing/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_billing_budget", &resource.Sweeper{
		Name: "scaleway_billing_budget",
		F:    testSweepBillingBudget,
	})
	resource.AddTestSweepers("scaleway_billing_budget_alert", &resource.Sweeper{
		Name: "scaleway_billing_budget_alert",
		F:    testSweepBillingBudgetAlert,
	})
	resource.AddTestSweepers("scaleway_billing_budget_alert_notification", &resource.Sweeper{
		Name: "scaleway_billing_budget_alert_notification",
		F:    testSweepBillingBudgetAlertNotification,
	})
}

func testSweepBillingBudget(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		api := billingSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the billing budgets")

		orgID, exists := scwClient.GetDefaultOrganizationID()
		if !exists {
			return errors.New("missing organizationID")
		}

		listBudgets, err := api.ListBudgets(&billingSDK.ListBudgetsRequest{
			OrganizationID: &orgID,
		})
		if err != nil {
			return fmt.Errorf("failed to list budgets: %w", err)
		}

		for _, budget := range listBudgets.Budgets {
			err = api.DeleteBudget(&billingSDK.DeleteBudgetRequest{
				BudgetID: budget.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete budget: %w", err)
			}
		}

		return nil
	})
}

func testSweepBillingBudgetAlert(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		api := billingSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the billing budget alerts")

		orgID, exists := scwClient.GetDefaultOrganizationID()
		if !exists {
			return errors.New("missing organizationID")
		}

		listBudgets, err := api.ListBudgets(&billingSDK.ListBudgetsRequest{
			OrganizationID: &orgID,
		})
		if err != nil {
			return fmt.Errorf("failed to list budgets: %w", err)
		}

		for _, budget := range listBudgets.Budgets {
			for _, alert := range budget.Alerts {
				err = api.DeleteBudgetAlert(&billingSDK.DeleteBudgetAlertRequest{
					BudgetAlertID: alert.ID,
				})
				if err != nil {
					return fmt.Errorf("failed to delete budget alert: %w", err)
				}
			}
		}

		return nil
	})
}

func testSweepBillingBudgetAlertNotification(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		api := billingSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the billing budget alert notifications")

		orgID, exists := scwClient.GetDefaultOrganizationID()
		if !exists {
			return errors.New("missing organizationID")
		}

		listBudgets, err := api.ListBudgets(&billingSDK.ListBudgetsRequest{
			OrganizationID: &orgID,
		})
		if err != nil {
			return fmt.Errorf("failed to list budgets: %w", err)
		}

		for _, budget := range listBudgets.Budgets {
			for _, alert := range budget.Alerts {
				for _, notification := range alert.Notifications {
					err = api.DeleteBudgetAlertNotification(&billingSDK.DeleteBudgetAlertNotificationRequest{
						BudgetAlertNotificationID: notification.ID,
					})
					if err != nil {
						return fmt.Errorf("failed to delete budget alert notification: %w", err)
					}
				}
			}
		}

		return nil
	})
}
