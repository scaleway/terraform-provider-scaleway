package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	cockpitSDK "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/cockpit"
)

func TestAccCockpitContactPoint_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resourceName := "scaleway_cockpit_contact_point.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isContactPointDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccCockpitContactPointConfig("initial@example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCockpitContactPointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", "initial@example.com"),
				),
			},
			{
				Config: testAccCockpitContactPointConfig("updated@example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCockpitContactPointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", "updated@example.com"),
				),
			},
		},
	})
}

func testAccCheckCockpitContactPointExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("contact point not found: %s", resourceName)
		}
		return nil
	}
}

func isContactPointDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_cockpit_contact_point" {
				continue
			}

			api, region, _, err := cockpit.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			projectID := rs.Primary.Attributes["project_id"]
			email := rs.Primary.Attributes["email"]

			contactPoints, err := api.ListContactPoints(&cockpitSDK.RegionalAPIListContactPointsRequest{
				Region:    region,
				ProjectID: projectID,
			})
			if err != nil {
				if !httperrors.Is404(err) {
					return fmt.Errorf("error retrieving contact points for project %s: %s", projectID, err)
				}
				continue
			}

			for _, cp := range contactPoints.ContactPoints {
				if cp.Email != nil && cp.Email.To == email {
					return fmt.Errorf("contact point with email %s still exists in project %s", email, projectID)
				}
			}
		}

		return nil
	}
}

func testAccCockpitContactPointConfig(email string) string {
	return fmt.Sprintf(`
		resource "scaleway_account_project" "project" {
						name = "tf_test_project"
		}
		
		resource "scaleway_cockpit_alert_manager" "manager" {
				project_id = scaleway_account_project.project.id
				enable     = true
		}

		resource "scaleway_cockpit_contact_point" "test" {
			project_id = scaleway_account_project.project.id
			email = "%s"
  			depends_on = [scaleway_cockpit_alert_manager.manager]
		}
	`, email)
}
