package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func TestAccCockpitAlertManager_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isAlertManagerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_test_project"
					}

					resource "scaleway_cockpit_alert_manager" "alert_manager" {
						project_id = scaleway_account_project.project.id
						enable     = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit_alert_manager.alert_manager", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable", "true"),
				),
			},
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_test_project"
					}

					resource "scaleway_cockpit_alert_manager" "alert_manager" {
						project_id = scaleway_account_project.project.id
						enable     = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit_alert_manager.alert_manager", "enable", "false"),
				),
			},
		},
	})
}

func isAlertManagerDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_cockpit_alert_manager" {
				continue
			}

			api := cockpit.NewRegionalAPI(meta.ExtractScwClient(tt.Meta))
			projectID := rs.Primary.Attributes["project_id"]

			alertManager, err := api.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
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
