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

func TestAccDeployment_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestClickHouseVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDeploymentDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_datawarehouse_deployment" "main" {
  name           = "tf-test-deploy-basic"
  version        = "%s"
  replica_count  = 1
  cpu_min        = 2
  cpu_max        = 4
  ram_per_cpu    = 4
  password       = "password@1234567"
}
`, latestVersion),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_datawarehouse_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "name", "tf-test-deploy-basic"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "version", latestVersion),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "replica_count", "1"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "cpu_min", "2"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "cpu_max", "4"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "ram_per_cpu", "4"),

					// Public endpoint is present
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "public_network.#", "1"),
				),
			},

			{
				// Update tags only
				Config: fmt.Sprintf(`
resource "scaleway_datawarehouse_deployment" "main" {
  name           = "tf-test-deploy-basic"
  version        = "%s"
  replica_count  = 1
  cpu_min        = 2
  cpu_max        = 4
  ram_per_cpu    = 4
  tags           = ["tag1", "tag2"]
  password       = "password@1234567"
}
`, latestVersion),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_datawarehouse_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "tags.1", "tag2"),

					// Public network still present
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "public_network.#", "1"),
				),
			},
		},
	})
}

func TestAccDeployment_WithPrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestClickHouseVersion(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDeploymentDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_vpc" "main" {
  region = "fr-par"
  name   = "TestAccDeployment_WithPrivateNetwork"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my_private_network"
  region = "fr-par"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_datawarehouse_deployment" "main" {
  name           = "tf-test-deploy-private-network"
  version        = "%s"
  replica_count  = 1
  cpu_min        = 2
  cpu_max        = 4
  ram_per_cpu    = 4
  password       = "password@1234567"

  private_network {
    pn_id = scaleway_vpc_private_network.pn01.id
  }
}
`, latestVersion),
				Check: resource.ComposeTestCheckFunc(
					isDeploymentPresent(tt, "scaleway_datawarehouse_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "name", "tf-test-deploy-private-network"),
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_datawarehouse_deployment.main", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_datawarehouse_deployment.main", "private_network.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_datawarehouse_deployment.main", "private_network.0.dns_record"),

					// Public endpoint is still present
					resource.TestCheckResourceAttr("scaleway_datawarehouse_deployment.main", "public_network.#", "1"),
				),
			},
		},
	})
}

func isDeploymentDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_datawarehouse_deployment" {
				continue
			}

			api, region, id, err := datawarehouse.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetDeployment(&datawarehouseSDK.GetDeploymentRequest{
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
