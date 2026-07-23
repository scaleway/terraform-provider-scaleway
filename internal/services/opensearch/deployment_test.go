package opensearch_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	searchdbSDK "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/opensearch"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
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
  node_count = 1
  node_type   = "%s"
  user_name   = "%s"
  password    = "ThisIsASecurePassword123!"
  volume {
    type        = "sbs_5k"
    size_in_gb = 5
  }
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "name", "tf-test-opensearch-basic"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "version", latestVersion),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "node_count", "1"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "node_type", nodeType),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "volume.0.type", "sbs_5k"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "volume.0.size_in_gb", "5"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.main", "endpoints.#", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_opensearch_deployment" "main" {
  name        = "tf-test-opensearch-basic"
  version     = "%s"
  node_count = 1
  node_type   = "%s"
  user_name   = "%s"
  password    = "ThisIsASecurePassword123!"
  tags        = ["tag1", "tag2"]
  volume {
    type        = "sbs_5k"
    size_in_gb = 5
  }
}
`, latestVersion, nodeType, deploymentTestUserName),
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
		CheckDestroy: resource.ComposeTestCheckFunc(
			isDeploymentDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
		),
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
  node_count = 1
  node_type   = "%s"
  user_name   = "%s"
  password    = "ThisIsASecurePassword123!"

  depends_on = [scaleway_vpc_private_network.main]

  volume {
    type        = "sbs_5k"
    size_in_gb = 5
  }
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.pn"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.pn", "name", "tf-test-opensearch-pn"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.pn", "endpoints.#", "1"),
					testAccCheckOpenSearchHasNoPrivateNetworkEndpoint("scaleway_opensearch_deployment.pn"),
					testAccCheckOpenSearchAPIHasNoPrivateEndpoint(tt, "scaleway_opensearch_deployment.pn"),
					resource.TestCheckResourceAttrSet("scaleway_opensearch_deployment.pn", "public_dashboard_url"),
				),
			},
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
  node_count = 1
  node_type   = "%s"
  user_name   = "%s"
  password    = "ThisIsASecurePassword123!"

  depends_on = [scaleway_vpc_private_network.main]

  private_network {
    private_network_id = scaleway_vpc_private_network.main.id
  }

  volume {
    type        = "sbs_5k"
    size_in_gb = 5
  }
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.pn"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.pn", "endpoints.#", "1"),
					testAccCheckOpenSearchHasPrivateNetworkEndpoint("scaleway_opensearch_deployment.pn"),
					testAccCheckOpenSearchAPIHasPrivateEndpoint(tt, "scaleway_opensearch_deployment.pn"),
					resource.TestCheckResourceAttrSet("scaleway_opensearch_deployment.pn", "public_dashboard_url"),
				),
			},
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
  node_count = 1
  node_type   = "%s"
  user_name   = "%s"
  password    = "ThisIsASecurePassword123!"

  depends_on = [scaleway_vpc_private_network.main]

  volume {
    type        = "sbs_5k"
    size_in_gb = 5
  }
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.pn"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.pn", "endpoints.#", "1"),
					testAccCheckOpenSearchHasNoPrivateNetworkEndpoint("scaleway_opensearch_deployment.pn"),
					testAccCheckOpenSearchAPIHasNoPrivateEndpoint(tt, "scaleway_opensearch_deployment.pn"),
					resource.TestCheckResourceAttrSet("scaleway_opensearch_deployment.pn", "public_dashboard_url"),
				),
			},
		},
	})
}

// TestAccDeployment_UpdatePrivateNetwork verifies Bug #2: when changing the
// private_network_id from one private network to another, the old private
// endpoint should be deleted and the new one created without a conflict error.
//
// The bug: updating only the private_network_id resulted in a conflict error
// saying that a private endpoint already exists, instead of replacing the old
// endpoint with the new one.
func TestAccDeployment_UpdatePrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestVersion(tt)
	nodeType := fetchAvailableNodeType(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isDeploymentDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_vpc" "main" {
  name = "tf-test-opensearch-vpc-update"
}

resource "scaleway_vpc_private_network" "pn1" {
  name   = "tf-test-opensearch-pn1"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_vpc_private_network" "pn2" {
  name   = "tf-test-opensearch-pn2"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_opensearch_deployment" "pn" {
  name        = "tf-test-opensearch-pn-update"
  version     = "%s"
  node_count  = 1
  node_type   = "%s"
  user_name   = "%s"
  password    = "ThisIsASecurePassword123!"

  depends_on = [scaleway_vpc_private_network.pn1, scaleway_vpc_private_network.pn2]

  private_network {
    private_network_id = scaleway_vpc_private_network.pn1.id
  }

  volume {
    type        = "sbs_5k"
    size_in_gb  = 5
  }
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.pn"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.pn", "endpoints.#", "1"),
					testAccCheckOpenSearchHasPrivateNetworkEndpoint("scaleway_opensearch_deployment.pn"),
					testAccCheckOpenSearchAPIPrivateNetworkID(tt, "scaleway_opensearch_deployment.pn", "scaleway_vpc_private_network.pn1"),
					resource.TestCheckResourceAttrSet("scaleway_opensearch_deployment.pn", "public_dashboard_url"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_vpc" "main" {
  name = "tf-test-opensearch-vpc-update"
}

resource "scaleway_vpc_private_network" "pn1" {
  name   = "tf-test-opensearch-pn1"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_vpc_private_network" "pn2" {
  name   = "tf-test-opensearch-pn2"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_opensearch_deployment" "pn" {
  name        = "tf-test-opensearch-pn-update"
  version     = "%s"
  node_count  = 1
  node_type   = "%s"
  user_name   = "%s"
  password    = "ThisIsASecurePassword123!"

  depends_on = [scaleway_vpc_private_network.pn1, scaleway_vpc_private_network.pn2]

  private_network {
    private_network_id = scaleway_vpc_private_network.pn2.id
  }

  volume {
    type        = "sbs_5k"
    size_in_gb  = 5
  }
}
`, latestVersion, nodeType, deploymentTestUserName),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_opensearch_deployment.pn"),
					resource.TestCheckResourceAttr("scaleway_opensearch_deployment.pn", "endpoints.#", "1"),
					testAccCheckOpenSearchHasPrivateNetworkEndpoint("scaleway_opensearch_deployment.pn"),
					testAccCheckOpenSearchAPIPrivateNetworkID(tt, "scaleway_opensearch_deployment.pn", "scaleway_vpc_private_network.pn2"),
					resource.TestCheckResourceAttrSet("scaleway_opensearch_deployment.pn", "public_dashboard_url"),
				),
			},
		},
	})
}

func testAccCheckOpenSearchHasPrivateNetworkEndpoint(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		n, err := strconv.Atoi(rs.Primary.Attributes["endpoints.#"])
		if err != nil {
			return fmt.Errorf("parse endpoints.#: %w", err)
		}

		for i := range n {
			if rs.Primary.Attributes[fmt.Sprintf("endpoints.%d.public", i)] == "false" &&
				rs.Primary.Attributes[fmt.Sprintf("endpoints.%d.private_network_id", i)] != "" {
				return nil
			}
		}

		return fmt.Errorf("expected a private network endpoint among %d endpoints", n)
	}
}

func testAccCheckOpenSearchHasNoPrivateNetworkEndpoint(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		n, err := strconv.Atoi(rs.Primary.Attributes["endpoints.#"])
		if err != nil {
			return fmt.Errorf("parse endpoints.#: %w", err)
		}

		for i := range n {
			if rs.Primary.Attributes[fmt.Sprintf("endpoints.%d.public", i)] == "false" &&
				rs.Primary.Attributes[fmt.Sprintf("endpoints.%d.private_network_id", i)] != "" {
				return fmt.Errorf("unexpected private network endpoint among %d endpoints", n)
			}
		}

		return nil
	}
}

// testAccCheckOpenSearchAPIHasPrivateEndpoint calls GetDeployment and verifies
// that the API actually has a private network endpoint.
//
// This is necessary because setDeploymentState filters endpoints in Terraform
// state to match the config, so state-based checks cannot detect API drift
// where both public and private endpoints coexist.
func testAccCheckOpenSearchAPIHasPrivateEndpoint(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		api, region, id, err := opensearch.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		var deployment *searchdbSDK.Deployment

		err = transport.RetryOn403(tt.T.Context(), func() error {
			deployment, err = api.GetDeployment(&searchdbSDK.GetDeploymentRequest{
				Region:       region,
				DeploymentID: id,
			}, scw.WithContext(tt.T.Context()))

			return err
		})
		if err != nil {
			return err
		}

		for _, ep := range deployment.Endpoints {
			if ep != nil && ep.PrivateNetwork != nil {
				return nil
			}
		}

		return fmt.Errorf("expected a private network endpoint in API state, got %d endpoints", len(deployment.Endpoints))
	}
}

// testAccCheckOpenSearchAPIHasNoPrivateEndpoint calls GetDeployment and verifies
// that the API has no private network endpoint (i.e., only public endpoints remain).
//
// This catches Bug #1: when removing a private_network block, the API may still
// have a stale private endpoint even though Terraform state was filtered to hide it.
func testAccCheckOpenSearchAPIHasNoPrivateEndpoint(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		api, region, id, err := opensearch.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		var deployment *searchdbSDK.Deployment

		err = transport.RetryOn403(tt.T.Context(), func() error {
			deployment, err = api.GetDeployment(&searchdbSDK.GetDeploymentRequest{
				Region:       region,
				DeploymentID: id,
			}, scw.WithContext(tt.T.Context()))

			return err
		})
		if err != nil {
			return err
		}

		for _, ep := range deployment.Endpoints {
			if ep != nil && ep.PrivateNetwork != nil {
				return fmt.Errorf("unexpected private network endpoint %s in API state", ep.ID)
			}
		}

		return nil
	}
}

// testAccCheckOpenSearchAPIPrivateNetworkID calls GetDeployment and verifies
// that the API has a private network endpoint with the expected private network ID.
//
// This catches Bug #2: when changing the private_network_id, the old endpoint
// should be deleted and the new one created without a conflict error. This check
// also ensures no stale private endpoint from the previous PN remains.
func testAccCheckOpenSearchAPIPrivateNetworkID(tt *acctest.TestTools, resourceName, pnResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		pnRS, ok := s.RootModule().Resources[pnResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", pnResourceName)
		}

		expectedPNID := locality.ExpandID(pnRS.Primary.ID)

		api, region, id, err := opensearch.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		var deployment *searchdbSDK.Deployment

		err = transport.RetryOn403(tt.T.Context(), func() error {
			deployment, err = api.GetDeployment(&searchdbSDK.GetDeploymentRequest{
				Region:       region,
				DeploymentID: id,
			}, scw.WithContext(tt.T.Context()))

			return err
		})
		if err != nil {
			return err
		}

		foundMatch := false

		for _, ep := range deployment.Endpoints {
			if ep == nil || ep.PrivateNetwork == nil {
				continue
			}

			if ep.PrivateNetwork.PrivateNetworkID == expectedPNID {
				foundMatch = true

				continue
			}

			return fmt.Errorf("unexpected private network endpoint with PN ID %s (expected %s)",
				ep.PrivateNetwork.PrivateNetworkID, expectedPNID)
		}

		if !foundMatch {
			return fmt.Errorf("expected a private network endpoint with PN ID %s in API state, got %d endpoints",
				expectedPNID, len(deployment.Endpoints))
		}

		return nil
	}
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

			err = transport.RetryOn403(tt.T.Context(), func() error {
				_, err = api.GetDeployment(&searchdbSDK.GetDeploymentRequest{
					Region:       region,
					DeploymentID: id,
				}, scw.WithContext(tt.T.Context()))

				return err
			})
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
