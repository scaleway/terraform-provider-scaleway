package datawarehouse_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	datawarehouseSDK "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/datawarehouse"
)

func TestAccUser_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestClickHouseVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isUserDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_datawarehouse_deployment" "test_deploy" {
  name           = "tf-test-deploy-user"
  version        = "%s"
  replica_count  = 1
  cpu_min        = 2
  cpu_max        = 4
  ram_per_cpu    = 4
  password       = "password@1234567"
}

resource "scaleway_datawarehouse_user" "test_user" {
  deployment_id = scaleway_datawarehouse_deployment.test_deploy.id
  name          = "tf_test_user"
  password      = "userPassword@123"
  is_admin      = false
}
`, latestVersion),
				Check: resource.ComposeTestCheckFunc(
					isUserPresent(tt, "scaleway_datawarehouse_user.test_user"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_user.test_user", "name", "tf_test_user"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_user.test_user", "is_admin", "false"),
				),
			},
			{
				// Update is_admin to true
				Config: fmt.Sprintf(`
resource "scaleway_datawarehouse_deployment" "test_deploy" {
  name           = "tf-test-deploy-user"
  version        = "%s"
  replica_count  = 1
  cpu_min        = 2
  cpu_max        = 4
  ram_per_cpu    = 4
  password       = "password@1234567"
}

resource "scaleway_datawarehouse_user" "test_user" {
  deployment_id = scaleway_datawarehouse_deployment.test_deploy.id
  name          = "tf_test_user"
  password      = "userPassword@123"
  is_admin      = true
}
`, latestVersion),
				Check: resource.ComposeTestCheckFunc(
					isUserPresent(tt, "scaleway_datawarehouse_user.test_user"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_user.test_user", "is_admin", "true"),
				),
			},
		},
	})
}

func isUserPresent(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID // format: region/deployment_id/name

		region, deploymentID, userName, err := datawarehouse.ResourceUserParseID(id)
		if err != nil {
			return fmt.Errorf("unexpected ID format (%s), expected region/deployment_id/name", id)
		}

		api := datawarehouse.NewAPI(tt.Meta)

		resp, err := api.ListUsers(&datawarehouseSDK.ListUsersRequest{
			Region:       region,
			DeploymentID: deploymentID,
			Name:         new(userName),
		}, scw.WithContext(context.Background()))
		if err != nil {
			return err
		}

		for _, u := range resp.Users {
			if u.Name == userName {
				return nil
			}
		}

		return fmt.Errorf("user %q not found in deployment %s", userName, deploymentID)
	}
}

func isUserDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_datawarehouse_user" {
				continue
			}

			id := rs.Primary.ID // format: region/deployment_id/name

			region, deploymentID, userName, err := datawarehouse.ResourceUserParseID(id)
			if err != nil {
				return fmt.Errorf("unexpected ID format (%s), expected region/deployment_id/name", id)
			}

			api := datawarehouse.NewAPI(tt.Meta)

			resp, err := api.ListUsers(&datawarehouseSDK.ListUsersRequest{
				Region:       region,
				DeploymentID: deploymentID,
				Name:         new(userName),
			}, scw.WithContext(context.Background()))

			if err != nil && !httperrors.Is404(err) {
				return err
			}
			// If no error, check if user still exists
			if err == nil {
				for _, u := range resp.Users {
					if u.Name == userName {
						return fmt.Errorf("user %s still exists after destroy", userName)
					}
				}
			}
		}

		return nil
	}
}
