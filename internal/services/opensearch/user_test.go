package opensearch_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	searchdbSDK "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/opensearch"
)

func TestAccUser_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestVersion(tt)
	nodeType := fetchAvailableNodeType(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isUserDestroyed(tt),
			isDeploymentDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_opensearch_deployment" "main" {
  name       = "tf-test-opensearch-user"
  version    = "%s"
  node_count = 1
  node_type  = "%s"
  user_name  = "%s"
  password   = "ThisIsASecurePassword123!"

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}

resource "scaleway_opensearch_user" "main" {
  deployment_id = scaleway_opensearch_deployment.main.id
  name          = "app_user"
  password      = "UserSecurePassword123!"
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isUserPresent(tt, "scaleway_opensearch_user.main"),
					resource.TestCheckResourceAttr("scaleway_opensearch_user.main", "name", "app_user"),
					resource.TestCheckResourceAttrSet("scaleway_opensearch_user.main", "deployment_id"),
					resource.TestCheckResourceAttrSet("scaleway_opensearch_user.main", "region"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_opensearch_deployment" "main" {
  name       = "tf-test-opensearch-user"
  version    = "%s"
  node_count = 1
  node_type  = "%s"
  user_name  = "%s"
  password   = "ThisIsASecurePassword123!"

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}

resource "scaleway_opensearch_user" "main" {
  deployment_id = scaleway_opensearch_deployment.main.id
  name          = "app_user"
  password      = "UserSecurePassword456!"
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isUserPresent(tt, "scaleway_opensearch_user.main"),
					resource.TestCheckResourceAttr("scaleway_opensearch_user.main", "name", "app_user"),
					resource.TestCheckResourceAttr("scaleway_opensearch_user.main", "password", "UserSecurePassword456!"),
				),
			},
			{
				ResourceName:            "scaleway_opensearch_user.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "password_wo", "password_wo_version"},
			},
		},
	})
}

func isUserPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		region, deploymentID, userName, err := opensearch.ResourceUserParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		api := opensearch.NewAPI(tt.Meta)

		res, err := api.ListUsers(&searchdbSDK.ListUsersRequest{
			Region:       region,
			DeploymentID: deploymentID,
			Name:         &userName,
		}, scw.WithContext(tt.T.Context()))
		if err != nil {
			return err
		}

		for _, u := range res.Users {
			if u.Username == userName {
				return nil
			}
		}

		return fmt.Errorf("user %s not found on deployment %s", userName, deploymentID)
	}
}

func isUserDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_opensearch_user" {
				continue
			}

			region, deploymentID, userName, err := opensearch.ResourceUserParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			api := opensearch.NewAPI(tt.Meta)

			res, err := api.ListUsers(&searchdbSDK.ListUsersRequest{
				Region:       region,
				DeploymentID: deploymentID,
				Name:         &userName,
			}, scw.WithContext(tt.T.Context()))
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return err
			}

			for _, u := range res.Users {
				if u.Username == userName {
					return fmt.Errorf("user %s still exists on deployment %s", userName, deploymentID)
				}
			}
		}

		return nil
	}
}
