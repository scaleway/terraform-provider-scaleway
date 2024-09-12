package inference_test

import (
	"fmt"
	inferencetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference/testfuncs"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	inferenceSDK "github.com/scaleway/scaleway-sdk-go/api/inference/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference"
)

func TestAccDeployment_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      inferencetestfuncs.IsDeploymentDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_inference_deployment" "main" {
						name = "test-inferenceSDK-deployment-basic"
						node_type = "L4"
						model_name = "meta/llama-3.1-8b-instruct:fp8"
						endpoints {
							public_endpoint = true
						}
						accept_eula = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(tt, "scaleway_inference_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main", "name", "test-inferenceSDK-deployment-basic"),
				),
			},
		},
	})
}

func TestAccDeployment_Endpoint(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      inferencetestfuncs.IsDeploymentDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_private_network" "pn01" {
						name = "private-network-test-inference"
					}
					resource "scaleway_inference_deployment" "main" {
						name = "test-inferenceSDK-deployment-endpoint-private"
						node_type = "L4"
						model_name = "meta/llama-3.1-8b-instruct:fp8"
						endpoints {
							private_endpoint = "${scaleway_vpc_private_network.pn01.id}"
						}
						accept_eula = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(tt, "scaleway_inference_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main", "name", "test-inferenceSDK-deployment-basic"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc_private_network" "pn01" {
						name = "private-network-test-inference"
					}
					resource "scaleway_inference_deployment" "main" {
						name = "test-inferenceSDK-deployment-basic-endpoints-private-public"
						node_type = "L4"
						model_name = "meta/llama-3.1-8b-instruct:fp8"
						endpoints {
							private_endpoint = "${scaleway_vpc_private_network.pn01.id}"
							public_endpoint = true
						}
						accept_eula = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(tt, "scaleway_inference_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main", "name", "test-inferenceSDK-deployment-basic"),
				),
			},
		},
	})
}

func TestAccDeployment_MinSize(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      inferencetestfuncs.IsDeploymentDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_inference_deployment" "main_size" {
						name = "test-inferenceSDK-deployment-min-size"
						node_type = "L4"
						model_name = "meta/llama-3.1-8b-instruct:fp8"
						endpoints {
							public_endpoint = true
						}
						accept_eula = true
						min_size = 2
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(tt, "scaleway_inference_deployment.main_size"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main_size", "min_size", "2"),
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
