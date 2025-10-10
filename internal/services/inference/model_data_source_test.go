package inference_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	inferencetestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference/testfuncs"
)

func TestAccDataSourceModel_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	modelName := "mistral/pixtral-12b-2409:bf16"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				data "scaleway_inference_model" "my-model" {
					name = "%s"
				}

`, modelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(tt, "data.scaleway_inference_model.my-model"),
					resource.TestCheckResourceAttr("data.scaleway_inference_model.my-model", "name", modelName),
				),
			},
		},
	})
}

func TestAccDataSourceModel_Custom(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	modelName := "TestAccDataSourceModel_Custom"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             inferencetestfuncs.IsModelDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_inference_model" "test" {
						name = "%s"
						url = "%s"
					}
				
				`, modelName, modelURLCompatible),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(tt, "scaleway_inference_model.test"),
					resource.TestCheckResourceAttr("scaleway_inference_model.test", "name", modelName),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_inference_model" "test" {
						name = "%s"
						url = "%s"
					}
				
					data "scaleway_inference_model" "my-model" {
						name = "%s"
					}`, modelName, modelURLCompatible, modelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelExists(tt, "data.scaleway_inference_model.my-model"),
					resource.TestCheckResourceAttr("data.scaleway_inference_model.my-model", "name", modelName),
				),
			},
		},
	})
}
