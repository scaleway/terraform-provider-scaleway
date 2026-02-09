package opensearch_test

import (
	"context"
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

func TestAccDeployment_Basic(t *testing.T) {
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
  name        = "tf-test-opensearch-basic"
  version     = "%s"
  node_amount = 1
  node_type   = "%s"
  password    = "ThisIsASecurePassword123!"
  volume {
    type       = "sbs_5k"
    size_bytes = 5000000000
  }
}
`, latestVersion, nodeType),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "name", "tf-test-opensearch-basic"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "version", latestVersion),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "node_amount", "1"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "node_type", nodeType),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "volume.0.type", "sbs_5k"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "volume.0.size_bytes", "5000000000"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "endpoints.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_opensearch_deployment" "main" {
  name        = "tf-test-opensearch-basic"
  version     = "%s"
  node_amount = 1
  node_type   = "%s"
  password    = "ThisIsASecurePassword123!"
  tags        = ["tag1", "tag2"]
  volume {
    type       = "sbs_5k"
    size_bytes = 5000000000
  }
}
`, latestVersion, nodeType),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "tags.1", "tag2"),
				),
			},
		},
	})
}

func TestAccDeployment_WithPrivateNetwork(t *testing.T) {
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
resource "scaleway_vpc" "main" {
  name = "tf-test-opensearch-vpc"
}

resource "scaleway_vpc_private_network" "main" {
  name   = "tf-test-opensearch-pn"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_opensearch_deployment" "pn" {
  name        = "tf-test-opensearch-pn"
  version     = "%s"
  node_amount = 1
  node_type   = "%s"
  password    = "ThisIsASecurePassword123!"
  
  private_network {
    private_network_id = scaleway_vpc_private_network.main.id
  }
  
  volume {
    type       = "sbs_5k"
    size_bytes = 5000000000
  }
}
`, latestVersion, nodeType),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.pn"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.pn", "name", "tf-test-opensearch-pn"),
					resource.TestCheckResourceAttrSet("scaleway_opensearch_deployment.pn", "private_network.0.private_network_id"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.pn", "endpoints.#", "1"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.pn", "endpoints.0.public", "false"),
					resource.TestCheckResourceAttrSet("scaleway_opensearch_deployment.pn", "endpoints.0.private_network_id"),
				),
			},
		},
	})
}

func isDeploymentDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_opensearch_deployment" {
				continue
			}

			api, region, id, err := opensearch.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetDeployment(&searchdbSDK.GetDeploymentRequest{
				Region:       region,
				DeploymentID: id,
			}, scw.WithContext(context.Background()))
			if err == nil {
				return fmt.Errorf("deployment %s still exists", id)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
