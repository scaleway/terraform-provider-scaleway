package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func TestAccCockpitAlertManager_CreateWithSingleContact(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]string{"initial@example.com"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.0", "initial@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]string{"updated@example.com"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.0", "updated@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
		},
	})
}

func TestAccCockpitAlertManager_CreateWithMultipleContacts(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]string{"initial1@example.com", "initial2@example.com"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.0", "initial1@example.com"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.1", "initial2@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]string{"updated1@example.com", "updated2@example.com"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.0", "updated1@example.com"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.1", "updated2@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
		},
	})
}

func TestAccCockpitAlertManager_UpdateSingleContact(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]string{"notupdated@example.com", "initial1@example.com"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.0", "notupdated@example.com"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.1", "initial1@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]string{"notupdated@example.com", "updated1@example.com"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.0", "notupdated@example.com"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "emails.1", "updated1@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
		},
	})
}

func TestAccCockpitAlertManager_EnableDisable(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCockpitAlertManagerEnableConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckAlertManagerEnabled(tt, "scaleway_cockpit_alert_manager.alert_manager", true),
				),
			},
			{
				Config: testAccCockpitAlertManagerEnableConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "false"),
					testAccCheckAlertManagerEnabled(tt, "scaleway_cockpit_alert_manager.alert_manager", false),
				),
			},
		},
	})
}

func testAccCockpitAlertManagerConfigWithContacts(emails []string) string {
	emailsConfig := "["
	for _, email := range emails {
		emailsConfig += fmt.Sprintf(`"%s", `, email)
	}
	if len(emails) > 0 {
		emailsConfig = emailsConfig[:len(emailsConfig)-2] // Remove the last comma and space
	}
	emailsConfig += "]"

	return fmt.Sprintf(`
		resource "scaleway_account_project" "project" {
			name = "tf_test_project"
		}

		resource "scaleway_cockpit_alert_manager" "alert_manager" {
			project_id = scaleway_account_project.project.id
			enable_managed_alerts     = true
			emails     = %s
		}
	`, emailsConfig)
}

func testAccCockpitAlertManagerEnableConfig(enable bool) string {
	return fmt.Sprintf(`
        resource "scaleway_account_project" "project" {
            name = "tf_test_project"
        }

        resource "scaleway_cockpit_alert_manager" "alert_manager" {
            project_id = scaleway_account_project.project.id
            enable_managed_alerts     = %t
        }
    `, enable)
}

func testAccCheckAlertManagerEnabled(tt *acctest.TestTools, resourceName string, expectedEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("alert manager not found: %s", resourceName)
		}

		api := cockpit.NewRegionalAPI(meta.ExtractScwClient(tt.Meta))
		projectID := rs.Primary.Attributes["project_id"]

		alertManager, err := api.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
			ProjectID: projectID,
		})
		if err != nil {
			return err
		}
		if alertManager.ManagedAlertsEnabled != expectedEnabled {
			return fmt.Errorf("alert manager enabled state %t does not match expected state %t", alertManager.AlertManagerEnabled, expectedEnabled)
		}
		return nil
	}
}

func testAccCheckCockpitContactPointExists(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("alert manager not found: %s", resourceName)
		}

		api := cockpit.NewRegionalAPI(meta.ExtractScwClient(tt.Meta))
		projectID := rs.Primary.Attributes["project_id"]

		contactPoints, err := api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
			ProjectID: projectID,
		})
		if err != nil {
			return err
		}
		for _, cp := range contactPoints.ContactPoints {
			if cp.Email != nil && cp.Email.To == rs.Primary.Attributes["emails.0"] {
				return nil
			}
		}
		return fmt.Errorf("contact point with email %s not found in project %s", rs.Primary.Attributes["emails.0"], projectID)
	}
}

func testAccCockpitAlertManagerAndContactsDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_cockpit_alert_manager" {
				continue
			}

			api := cockpit.NewRegionalAPI(meta.ExtractScwClient(tt.Meta))
			projectID := rs.Primary.Attributes["project_id"]
			region := scw.RegionFrPar
			alertManager, err := api.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
				Region:    region,
				ProjectID: projectID,
			})

			if !httperrors.Is404(err) && !httperrors.Is403(err) {
				return err
			}

			if alertManager == nil {
				return nil
			}
			if alertManager.AlertManagerEnabled {
				return fmt.Errorf("cockpit alert manager (%s) is still enabled", rs.Primary.ID)
			}
		}
		return nil
	}
}
