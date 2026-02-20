package inference_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	inferenceSDK "github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference"
	inferencetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference/testfuncs"
)

const (
	modelURLCompatible = "https://huggingface.co/agentica-org/DeepCoder-14B-Preview"
	nodeTypeH100       = "H100"
)

func TestAccModel_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	modelName := "TestAccModel_Basic"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             inferencetestfuncs.IsModelDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_inference_model" "test" {
						name = "%s"
						url = "%s"
					}`, modelName, modelURLCompatible),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(tt, "scaleway_inference_model.test"),
					resource.TestCheckResourceAttr("scaleway_inference_model.test", "name", modelName),
				),
			},
		},
	})
}

func TestAccModel_WithSecretWO(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	modelName := "TestAccModel_WithSecretWO"
	secretValue := "test-secret-token"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             inferencetestfuncs.IsModelDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_inference_model" "test" {
						name = "%s"
						url = "%s"
						secret_wo = "%s"
						secret_wo_version = 1
					}`, modelName, modelURLCompatible, secretValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(tt, "scaleway_inference_model.test"),
					resource.TestCheckResourceAttr("scaleway_inference_model.test", "name", modelName),
					// Verify that secret_wo is not stored in state
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["scaleway_inference_model.test"]
						if !ok {
							return errors.New("can't find model resource")
						}

						if _, ok := rs.Primary.Attributes["secret_wo"]; ok {
							return errors.New("secret_wo should not be stored in state")
						}

						if _, ok := rs.Primary.Attributes["secret"]; ok {
							return errors.New("secret should not be stored in state")
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccModel_DeployModelOnServer(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	modelName := "TestAccModel_DeployModelOnServer"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             inferencetestfuncs.IsModelDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_inference_model" "test" {
						name = "%s"
						url = "%s"
					}
					resource "scaleway_inference_deployment" "main" {
						name = "%s"
						node_type = "%s"
						model_id = scaleway_inference_model.test.id
  						public_endpoint {
    						is_enabled = true
 		 				}
						accept_eula = true
					}
				`, modelName, modelURLCompatible, modelName, nodeTypeH100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(tt, "scaleway_inference_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_inference_deployment.main", "model_name", modelName),
				),
			},
		},
	})
}

func testAccCheckModelExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find model resource name: %s", n)
		}

		api, region, id, err := inference.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetModel(&inferenceSDK.GetModelRequest{
			Region:  region,
			ModelID: id,
		})

		return err
	}
}
