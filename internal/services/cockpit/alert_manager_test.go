package cockpit_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]map[string]string{
					{"email": "initial@example.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]map[string]string{
					{"email": "initial1@example.com"},
					{"email": "initial2@example.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCockpitAlertManagerConfigWithContacts([]map[string]string{
					{"email": "notupdated@example.com"},
					{"email": "initial1@example.com"},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
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

func TestAccCockpitAlertManager_LegacyManagedAlerts(t *testing.T) {
	t.Skip("TestAccCockpitAlertManager_LegacyManagedAlerts skipped: encountered HTTP 500 errors from the Scaleway Cockpit API.")

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_cockpit_alert_manager" "main" {
						enable_managed_alerts = true

						contact_points {
							email = "legacy@example.com"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.#", "0"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "contact_points.0.email", "legacy@example.com"),
					testAccCheckCockpitContactPointExists(tt, "scaleway_cockpit_alert_manager.main"),
					testAccCheckManagedAlertsEnabled(tt, "scaleway_cockpit_alert_manager.main", true),
				),
			},
		},
	})
}

func TestAccCockpitAlertManager_IDHandling(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_cockpit_alert_manager" "main" {
						contact_points {
							email = "test@example.com"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "alert_manager_url"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "contact_points.0.email", "test@example.com"),
					testAccCheckAlertManagerIDFormat(tt, "scaleway_cockpit_alert_manager.main"),
				),
			},
			{
				Config: `
					resource "scaleway_cockpit_alert_manager" "main" {
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

	var contactsConfigSb230 strings.Builder
	for _, contact := range contactPoints {
		contactsConfigSb230.WriteString(fmt.Sprintf(`
		contact_points {
			email = "%s"
		}`, contact["email"]))
	}

	contactsConfig += contactsConfigSb230.String()

	return fmt.Sprintf(`
		resource "scaleway_cockpit_alert_manager" "alert_manager" {
			%s
		}
	`, contactsConfig)
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
			return errors.New("alert manager ID should have 3 parts, got " + strconv.Itoa(len(parts)) + ": " + id)
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

func TestAccCockpitAlertManager_WithPreconfiguredAlerts(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_cockpit_alert_manager" "main" {
						# Enable 2 specific preconfigured alerts (stable IDs)
						preconfigured_alert_ids = [
							"6c6843af-1815-46df-9e52-6feafcf31fd7", # PostgreSQL Too Many Connections
							"eb8a941e-698d-47d6-b62d-4b6c13f7b4b7"  # MySQL Too Many Connections
						]

						contact_points {
							email = "test@example.com"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "region"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "alert_manager_url"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "contact_points.0.email", "test@example.com"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.*", "6c6843af-1815-46df-9e52-6feafcf31fd7"),
					resource.TestCheckTypeSetElemAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.*", "eb8a941e-698d-47d6-b62d-4b6c13f7b4b7"),
				),
			},
		},
	})
}

func TestAccCockpitAlertManager_UpdatePreconfiguredAlerts(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCockpitAlertManagerAndContactsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_cockpit_alert_manager" "main" {
						# Enable a specific PostgreSQL alert (stable ID)
						preconfigured_alert_ids = [
							"6c6843af-1815-46df-9e52-6feafcf31fd7" # PostgreSQL Too Many Connections
						]

						contact_points {
							email = "test@example.com"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.main", "project_id"),
					testAccCheckPreconfiguredAlertsCount(tt, "scaleway_cockpit_alert_manager.main", 1),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.*", "6c6843af-1815-46df-9e52-6feafcf31fd7"),
				),
			},
			{
				Config: `
					resource "scaleway_cockpit_alert_manager" "main" {
						# Enable 2 specific alerts (stable IDs)
						preconfigured_alert_ids = [
							"6c6843af-1815-46df-9e52-6feafcf31fd7", # PostgreSQL Too Many Connections
							"eb8a941e-698d-47d6-b62d-4b6c13f7b4b7"  # MySQL Too Many Connections
						]

						contact_points {
							email = "test@example.com"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreconfiguredAlertsCount(tt, "scaleway_cockpit_alert_manager.main", 2),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.*", "6c6843af-1815-46df-9e52-6feafcf31fd7"),
					resource.TestCheckTypeSetElemAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.*", "eb8a941e-698d-47d6-b62d-4b6c13f7b4b7"),
				),
			},
			{
				Config: `
					resource "scaleway_cockpit_alert_manager" "main" {
						# Disable all
						preconfigured_alert_ids = []

						contact_points {
							email = "test@example.com"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.main", "preconfigured_alert_ids.#", "0"),
					testAccCheckPreconfiguredAlertsCount(tt, "scaleway_cockpit_alert_manager.main", 0),
				),
			},
		},
	})
}

func testAccCheckPreconfiguredAlertsCount(tt *acctest.TestTools, resourceName string, expectedCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return errors.New("alert manager not found: " + resourceName)
		}

		actualCountStr := rs.Primary.Attributes["preconfigured_alert_ids.#"]

		actualCount, err := strconv.Atoi(actualCountStr)
		if err != nil {
			return fmt.Errorf("failed to parse preconfigured_alert_ids count: %w", err)
		}

		if actualCount != expectedCount {
			return fmt.Errorf("expected %d user-requested preconfigured alerts in state, got %d", expectedCount, actualCount)
		}

		api := cockpit.NewRegionalAPI(meta.ExtractScwClient(tt.Meta))
		projectID := rs.Primary.Attributes["project_id"]
		region := scw.Region(rs.Primary.Attributes["region"])

		userRequestedIDs := make(map[string]bool)

		for i := range actualCount {
			alertID := rs.Primary.Attributes[fmt.Sprintf("preconfigured_alert_ids.%d", i)]
			if alertID != "" {
				userRequestedIDs[alertID] = true
			}
		}

		alerts, err := api.ListAlerts(&cockpit.RegionalAPIListAlertsRequest{
			Region:          region,
			ProjectID:       projectID,
			IsPreconfigured: scw.BoolPtr(true),
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list alerts: %w", err)
		}

		enabledUserAlertsCount := 0

		for _, alert := range alerts.Alerts {
			if alert.PreconfiguredData != nil && alert.PreconfiguredData.PreconfiguredRuleID != "" {
				ruleID := alert.PreconfiguredData.PreconfiguredRuleID
				if userRequestedIDs[ruleID] {
					if alert.RuleStatus == cockpit.AlertStatusEnabled || alert.RuleStatus == cockpit.AlertStatusEnabling {
						enabledUserAlertsCount++
					}
				}
			}
		}

		if enabledUserAlertsCount != expectedCount {
			return fmt.Errorf("expected %d user-requested alerts to be enabled in API, got %d", expectedCount, enabledUserAlertsCount)
		}

		return nil
	}
}

func testAccCheckManagedAlertsEnabled(tt *acctest.TestTools, resourceName string, expectedEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return errors.New("alert manager not found: " + resourceName)
		}

		api := cockpit.NewRegionalAPI(meta.ExtractScwClient(tt.Meta))
		projectID := rs.Primary.Attributes["project_id"]
		region := scw.Region(rs.Primary.Attributes["region"])

		alertManager, err := api.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
			Region:    region,
			ProjectID: projectID,
		})
		if err != nil {
			return err
		}

		if alertManager.ManagedAlertsEnabled != expectedEnabled {
			return fmt.Errorf("expected managed alerts enabled state %t, got %t", expectedEnabled, alertManager.ManagedAlertsEnabled)
		}

		return nil
	}
}
