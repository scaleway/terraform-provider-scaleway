package messageq_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceMessageQDeployment_ByName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestVersion(tt)
	nodeType := fetchAvailableNodeType(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDeploymentDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_messageq_deployment" "main" {
  name       = "tf-test-ds-messageq-by-name"
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

data "scaleway_messageq_deployment" "by_name" {
  name = scaleway_messageq_deployment.main.name
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_messageq_deployment.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_messageq_deployment.by_name", "id",
						"scaleway_messageq_deployment.main", "id",
					),
					resource.TestCheckResourceAttr("data.scaleway_messageq_deployment.by_name", "name", "tf-test-ds-messageq-by-name"),
					resource.TestCheckResourceAttr("data.scaleway_messageq_deployment.by_name", "version", latestVersion),
					resource.TestCheckResourceAttr("data.scaleway_messageq_deployment.by_name", "node_count", "1"),
					resource.TestCheckResourceAttr("data.scaleway_messageq_deployment.by_name", "node_type", nodeType),
					resource.TestCheckResourceAttrSet("data.scaleway_messageq_deployment.by_name", "status"),
				),
			},
		},
	})
}

func TestAccDataSourceMessageQDeployment_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestVersion(tt)
	nodeType := fetchAvailableNodeType(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDeploymentDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_messageq_deployment" "main" {
  name       = "tf-test-ds-messageq-by-id"
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

data "scaleway_messageq_deployment" "by_id" {
  deployment_id = scaleway_messageq_deployment.main.id
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_messageq_deployment.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_messageq_deployment.by_id", "id",
						"scaleway_messageq_deployment.main", "id",
					),
					resource.TestCheckResourceAttr("data.scaleway_messageq_deployment.by_id", "name", "tf-test-ds-messageq-by-id"),
					resource.TestCheckResourceAttrSet("data.scaleway_messageq_deployment.by_id", "status"),
				),
			},
		},
	})
}
