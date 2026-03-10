package opensearch_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceOpenSearchDeployment_ByName(t *testing.T) {
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
resource "scaleway_opensearch_deployment" "main" {
  name        = "tf-test-ds-opensearch-by-name"
  version     = "%s"
  node_amount = 1
  node_type   = "%s"
  password    = "ThisIsASecurePassword123!"
  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}

data "scaleway_opensearch_deployment" "by_name" {
  name = scaleway_opensearch_deployment.main.name
}
`, latestVersion, nodeType),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_opensearch_deployment.by_name", "id",
						"scaleway_opensearch_deployment.main", "id",
					),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_name", "name", "tf-test-ds-opensearch-by-name"),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_name", "version", latestVersion),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_name", "node_amount", "1"),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_name", "node_type", nodeType),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_name", "volume.0.type", "sbs_5k"),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_name", "volume.0.size_in_gb", "5"),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_name", "endpoints.#", "1"),
					resource.TestCheckResourceAttrSet("data.scaleway_opensearch_deployment.by_name", "status"),
					resource.TestCheckResourceAttrSet("data.scaleway_opensearch_deployment.by_name", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceOpenSearchDeployment_ByID(t *testing.T) {
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
resource "scaleway_opensearch_deployment" "main" {
  name        = "tf-test-ds-opensearch-by-id"
  version     = "%s"
  node_amount = 1
  node_type   = "%s"
  password    = "ThisIsASecurePassword123!"
  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}

data "scaleway_opensearch_deployment" "by_id" {
  deployment_id = scaleway_opensearch_deployment.main.id
}
`, latestVersion, nodeType),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_opensearch_deployment.by_id", "id",
						"scaleway_opensearch_deployment.main", "id",
					),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_id", "name", "tf-test-ds-opensearch-by-id"),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_id", "version", latestVersion),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_id", "node_amount", "1"),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_id", "node_type", nodeType),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_id", "volume.0.type", "sbs_5k"),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_id", "volume.0.size_in_gb", "5"),
					resource.TestCheckResourceAttr("data.scaleway_opensearch_deployment.by_id", "endpoints.#", "1"),
					resource.TestCheckResourceAttrSet("data.scaleway_opensearch_deployment.by_id", "status"),
					resource.TestCheckResourceAttrSet("data.scaleway_opensearch_deployment.by_id", "created_at"),
				),
			},
		},
	})
}
