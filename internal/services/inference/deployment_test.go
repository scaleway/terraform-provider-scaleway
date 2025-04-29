package inference_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	inferenceSDK "github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference"
	inferencetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference/testfuncs"
)

const (
	modelID  = "ca90eec6-c8b2-4814-8b00-8bb3bc803879"
	nodeType = "L4"
)

func TestAccDeployment_Basic(t *testing.T) {
	t.Skip("Can not run tests at the moment")
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      inferencetestfuncs.IsDeploymentDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_inference_deployment" "main" {
						name = "test-inference-deployment-basic"
						node_type = "%s"
						model_id = "%s"
  						public_endpoint {
    						is_enabled = true
 		 				}
						accept_eula = true
					}
				`, nodeType, modelID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(tt, "scaleway_inference_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main", "name", "test-inference-deployment-basic"),
				),
			},
		},
	})
}

func TestAccDeployment_Endpoint(t *testing.T) {
	t.Skip("Can not run tests at the moment")
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      inferencetestfuncs.IsDeploymentDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "pn01" {
						name = "private-network-test-inference"
					}
					resource "scaleway_inference_deployment" "main" {
						name = "test-inference-deployment-endpoint-private"
						node_type = "%s"
						model_id = "%s"
						private_endpoint {
							private_network_id = "${scaleway_vpc_private_network.pn01.id}"
						}
						accept_eula = true
					}
				`, nodeType, modelID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(tt, "scaleway_inference_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main", "name", "test-inference-deployment-endpoint-private"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main", "node_type", "L4"),
					resource.TestCheckResourceAttrPair("scaleway_inference_deployment.main", "private_endpoint.0.private_network_id", "scaleway_vpc_private_network.pn01", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_vpc_private_network" "pn01" {
						name = "private-network-test-inference-public"
					}
					resource "scaleway_inference_deployment" "main" {
						name = "test-inference-deployment-basic-endpoints-private-public"
						node_type = "%s"
						model_id = "%s"
						private_endpoint {
							private_network_id = "${scaleway_vpc_private_network.pn01.id}"
						}
						public_endpoint {
							is_enabled = true
						}
						accept_eula = true
					}
				`, nodeType, modelID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(tt, "scaleway_inference_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main", "name", "test-inference-deployment-basic-endpoints-private-public"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main", "public_endpoint.0.is_enabled", "true"),
					resource.TestCheckResourceAttrPair("scaleway_inference_deployment.main", "private_endpoint.0.private_network_id", "scaleway_vpc_private_network.pn01", "id"),
				),
			},
		},
	})
}

func testAccCheckDeploymentExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := inference.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDeployment(&inferenceSDK.GetDeploymentRequest{
			DeploymentID: id,
			Region:       region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
