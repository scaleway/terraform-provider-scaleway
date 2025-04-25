package inference_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	inferenceSDK "github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference"
	inferencetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference/testfuncs"
	"regexp"
	"testing"
)

const (
	modelURLCompatible    = "https://huggingface.co/agentica-org/DeepCoder-14B-Preview"
	modelURLNotCompatible = "https://huggingface.co/google/gemma-3-4b-it"
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

func TestAccCustomModel_NotCompatible(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	modelName := "TestAccCustomModel_NotCompatible"

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
					}`, modelName, modelURLNotCompatible),
				ExpectError: regexp.MustCompile("scaleway-sdk-go: precondition failed: , the model with ID 'google/gemma-3-4b-it' is not supported. access to model google/gemma-3-4b-it is restricted. Check your permissions to access the repository at https://huggingface.co/google/gemma-3-4b-it and ensure your credentials are valid Please visit https://www.scaleway.com/en/docs/ai-data/managed-inference/reference-content/supported-models for more details about the supported models."),
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
		if err != nil {
			return err
		}
		return nil
	}
}
