package messageq_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	messageqSDK "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/messageq"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
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
resource "scaleway_messageq_deployment" "main" {
  name       = "tf-test-messageq-user"
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

resource "scaleway_messageq_user" "app" {
  deployment_id = scaleway_messageq_deployment.main.id
  name          = "app-user"
  password      = "AnotherSecurePassword123!"
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_messageq_deployment.main"),
					isUserPresent(tt, "scaleway_messageq_user.app"),
					resource.TestCheckResourceAttr("scaleway_messageq_user.app", "name", "app-user"),
					resource.TestCheckResourceAttrSet("scaleway_messageq_user.app", "deployment_id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_messageq_deployment" "main" {
  name       = "tf-test-messageq-user"
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

resource "scaleway_messageq_user" "app" {
  deployment_id = scaleway_messageq_deployment.main.id
  name          = "app-user"
  password      = "RotatedSecurePassword123!"
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isUserPresent(tt, "scaleway_messageq_user.app"),
					resource.TestCheckResourceAttr("scaleway_messageq_user.app", "password", "RotatedSecurePassword123!"),
				),
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

		idParts := identity.ParseMultiPartID(rs.Primary.ID, "region", "deployment_id", "name")
		region := scw.Region(idParts["region"])
		deploymentID := idParts["deployment_id"]
		userName := idParts["name"]

		api := messageq.NewAPI(tt.Meta)

		return transport.RetryOn403(tt.T.Context(), func() error {
			res, err := api.ListUsers(&messageqSDK.ListUsersRequest{
				Region:       region,
				DeploymentID: deploymentID,
				Name:         &userName,
			}, scw.WithContext(tt.T.Context()))
			if err != nil {
				return err
			}

			if len(res.Users) == 0 {
				return fmt.Errorf("user %s not found on deployment %s", userName, deploymentID)
			}

			return nil
		})
	}
}

func isUserDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_messageq_user" {
				continue
			}

			idParts := identity.ParseMultiPartID(rs.Primary.ID, "region", "deployment_id", "name")
			region := scw.Region(idParts["region"])
			deploymentID := idParts["deployment_id"]
			userName := idParts["name"]

			api := messageq.NewAPI(tt.Meta)

			err := transport.RetryOn403(tt.T.Context(), func() error {
				res, err := api.ListUsers(&messageqSDK.ListUsersRequest{
					Region:       region,
					DeploymentID: deploymentID,
					Name:         &userName,
				}, scw.WithContext(tt.T.Context()))
				if err != nil {
					return err
				}

				if len(res.Users) > 0 {
					return fmt.Errorf("user %s still exists", userName)
				}

				return nil
			})
			if err != nil && !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
