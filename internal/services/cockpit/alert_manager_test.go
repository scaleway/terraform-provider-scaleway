package cockpit_test

import (
	"errors"
	"fmt"
	"strings"
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
				Config: testAccCockpitAlertManagerConfigWithContacts([]map[string]string{
					{"email": "initial@example.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.0.email", "initial@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]map[string]string{
					{"email": "updated@example.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.0.email", "updated@example.com"),
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
				Config: testAccCockpitAlertManagerConfigWithContacts([]map[string]string{
					{"email": "initial1@example.com"},
					{"email": "initial2@example.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.0.email", "initial1@example.com"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.1.email", "initial2@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]map[string]string{
					{"email": "updated1@example.com"},
					{"email": "updated2@example.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.0.email", "updated1@example.com"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.1.email", "updated2@example.com"),
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
				Config: testAccCockpitAlertManagerConfigWithContacts([]map[string]string{
					{"email": "notupdated@example.com"},
					{"email": "initial1@example.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.0.email", "notupdated@example.com"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.1.email", "initial1@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]map[string]string{
					{"email": "notupdated@example.com"},
					{"email": "updated1@example.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.0.email", "notupdated@example.com"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "contact_points.1.email", "updated1@example.com"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "alert_manager_url"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.alert_manager"),
				),
			},
		},
	})
}

func TestAccCockpitAlertManager_EnableDisable(t *testing.T) {
	t.Skip("TestAccCockpit_WithSourceEndpoints skipped: encountered repeated HTTP 500 errors from the Scaleway Cockpit API.")

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

func TestAccCockpitAlertManager_IDHandling(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_test_cockpit_alert_manager_id"
					}

					resource "scaleway_cockpit_alert_manager" "main" {
						project_id = scaleway_account_project.project.id
						enable_managed_alerts = true

						contact_points {
							email = "test@example.com"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "region"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "enable_managed_alerts", "true"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "alert_manager_url"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "contact_points.0.email", "test@example.com"),
					testAccCheckAlertManagerIDFormat(tt, "scaleway_cockpit_alert_manager.main"),
				),
			},
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_test_cockpit_alert_manager_id"
					}

					resource "scaleway_cockpit_alert_manager" "main" {
						project_id = scaleway_account_project.project.id
						enable_managed_alerts = true

						contact_points {
							email = "updated@example.com"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "region"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "contact_points.0.email", "updated@example.com"),
					testAccCheckAlertManagerIDFormat(tt, "scaleway_cockpit_alert_manager.main"),
				),
			},
		},
	})
}

func testAccCockpitAlertManagerConfigWithContacts(contactPoints []map[string]string) string {
	contactsConfig := ""
	for _, contact := range contactPoints {
		contactsConfig += fmt.Sprintf(`
		contact_points {
			email = "%s"
		}`, contact["email"])
	}

	return fmt.Sprintf(`
		resource "scaleway_account_project" "project" {
			name = "tf_test_project"
		}

		resource "scaleway_cockpit_alert_manager" "alert_manager" {
			project_id = scaleway_account_project.project.id
			enable_managed_alerts = true
			%s
		}
	`, contactsConfig)
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
			return errors.New("alert manager not found: " + resourceName)
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
			return errors.New("alert manager not found: " + resourceName)
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
			if cp.Email != nil && cp.Email.To == rs.Primary.Attributes["contact_points.0.email"] {
				return nil
			}
		}

		return errors.New("contact point with email " + rs.Primary.Attributes["emails.0"] + " not found in project " + projectID)
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
				return errors.New("cockpit alert manager (" + rs.Primary.ID + ") is still enabled")
			}
		}

		return nil
	}
}

// testAccCheckAlertManagerIDFormat verifies the ID format
func testAccCheckAlertManagerIDFormat(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return errors.New("alert manager not found: " + resourceName)
		}

		id := rs.Primary.ID
		if id == "" {
			return errors.New("alert manager ID is empty")
		}

		parts := strings.Split(id, "/")
		if len(parts) != 3 {
			return errors.New("alert manager ID should have 3 parts, got " + fmt.Sprintf("%d", len(parts)) + ": " + id)
		}

		region := parts[0]
		projectID := parts[1]

		if region == "" {
			return errors.New("region part of ID is empty")
		}

		if projectID == "" {
			return errors.New("project ID part of ID is empty")
		}

		if parts[2] != "1" {
			return errors.New("third part of ID should be '1', got " + parts[2])
		}

		expectedProjectID := rs.Primary.Attributes["project_id"]
		if expectedProjectID != projectID {
			return errors.New("project_id in attributes (" + expectedProjectID + ") doesn't match project_id in ID (" + projectID + ")")
		}

		expectedRegion := rs.Primary.Attributes["region"]
		if expectedRegion != region {
			return errors.New("region in attributes (" + expectedRegion + ") doesn't match region in ID (" + region + ")")
		}

		return nil
	}
}
