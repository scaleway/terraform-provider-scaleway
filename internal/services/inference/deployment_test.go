package inference

import (
	"fmt"
	inference "github.com/scaleway/scaleway-sdk-go/api/inference/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDeployment_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckDeploymentDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_deployment main {
						name = "test-inference-deployment-basic"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(tt, "scaleway_deployment.main"),
					resource.TestCheckResourceAttr("scaleway_deployment.main", "name", "test-inference-deployment-basic"),
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

		api, region, id, err := newAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDeployment(&inference.GetDeploymentRequest{
			DeploymentID: id,
			Region:       region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckDeploymentDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_deployment" {
				continue
			}

			api, region, id, err := newAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteDeployment(&inference.DeleteDeploymentRequest{
				DeploymentID: id,
				Region:       region,
			})

			if err == nil {
				return fmt.Errorf("inference deployment (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
