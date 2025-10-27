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

	api := datawarehouse.NewAPI(tt.Meta)

	versionsResp, err := api.ListVersions(&datawarehouseSDK.ListVersionsRequest{}, scw.WithAllPages())
	if err != nil {
		t.Fatalf("unable to fetch datawarehouse versions: %s", err)
	}

	if len(versionsResp.Versions) == 0 {
		t.Fatal("no datawarehouse versions available")
	}

	latestVersion := versionsResp.Versions[0].Version

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
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

// Helpers

// isDeploymentPresent is now defined in helpers_test.go

func isDeploymentDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_datawarehouse_deployment" {
				continue
			}

			id := rs.Primary.ID
			region := rs.Primary.Attributes["region"]

			api := datawarehouse.NewAPI(tt.Meta)

			_, err := api.GetDeployment(&datawarehouseSDK.GetDeploymentRequest{
				Region:       scw.Region(region),
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
