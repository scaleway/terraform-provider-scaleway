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
	modelURLCompatible = "https://huggingface.co/agentica-org/DeepCoder-14B-Preview"
	nodeTypeH100       = "H100"
)

func TestAccCustomModel_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	modelName := "TestAccCustomModel_Basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      inferencetestfuncs.IsCustomModelDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_inference_custom_model" "test" {
						name = "%s"
						url = "%s"
					}`, modelName, modelURLCompatible),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomModelExists(tt, "scaleway_inference_custom_model.test"),
					resource.TestCheckResourceAttr("scaleway_inference_custom_model.test", "name", modelName),
				),
			},
		},
	})
}

func TestAccCustomModel_DeployModelOnServer(t *testing.T) {
	t.Skip("Can not run tests at the moment")

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	modelName := "TestAccCustomModel_DeployModelOnServer"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      inferencetestfuncs.IsCustomModelDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_inference_custom_model" "test" {
						name = "%s"
						url = "%s"
					}
					resource "scaleway_inference_deployment" "main" {
						name = "%s"
						node_type = "%s"
						model_id = scaleway_inference_custom_model.test.id
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

func testAccCheckCustomModelExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find custom model resource name: %s", n)
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
