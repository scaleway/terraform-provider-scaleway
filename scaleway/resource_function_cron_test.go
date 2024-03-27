package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func init() {
	resource.AddTestSweepers("scaleway_function_cron", &resource.Sweeper{
		Name: "scaleway_function_cron",
		F:    testSweepFunctionCron,
	})
}

func testSweepFunctionCron(_ string) error {
	return sweepRegions([]scw.Region{scw.RegionFrPar}, func(scwClient *scw.Client, region scw.Region) error {
		functionAPI := function.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the function cron in (%s)", region)
		listCron, err := functionAPI.ListCrons(
			&function.ListCronsRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing cron in (%s) in sweeper: %s", region, err)
		}

		for _, cron := range listCron.Crons {
			_, err := functionAPI.DeleteCron(&function.DeleteCronRequest{
				CronID: cron.ID,
				Region: region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting cron in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayFunctionCron_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionCronDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "tf-tests-function-cron-basic"
					}

					resource scaleway_function main {
						name = "tf-tests-cron-basic"
						namespace_id = scaleway_function_namespace.main.id
						runtime = "node14"
						privacy = "private"
						handler = "handler.handle"
					}

					resource scaleway_function_cron main {
						name = "tf-tests-cron-basic"
						function_id = scaleway_function.main.id
						schedule = "0 0 * * *"
						args = jsonencode({})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionCronExists(tt, "scaleway_function_cron.main"),
					resource.TestCheckResourceAttr("scaleway_function_cron.main", "schedule", "0 0 * * *"),
					resource.TestCheckResourceAttr("scaleway_function_cron.main", "name", "tf-tests-cron-basic"),
				),
			},
		},
	})
}

func TestAccScalewayFunctionCron_NameUpdate(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionCronDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "tf-tests-function-cron-name-update"
					}

					resource scaleway_function main {
						name = "tf-tests-function-cron-name-update"
						namespace_id = scaleway_function_namespace.main.id
						runtime = "node14"
						privacy = "private"
						handler = "handler.handle"
					}

					resource scaleway_function_cron main {
						name = "tf-tests-function-cron-name-update"
						function_id = scaleway_function.main.id
						schedule = "0 0 * * *"
						args = jsonencode({})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionCronExists(tt, "scaleway_function_cron.main"),
					resource.TestCheckResourceAttr("scaleway_function_cron.main", "schedule", "0 0 * * *"),
					resource.TestCheckResourceAttr("scaleway_function_cron.main", "name", "tf-tests-function-cron-name-update"),
				),
			},
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "tf-tests-function-cron-name-update"
					}

					resource scaleway_function main {
						name = "tf-tests-function-cron-name-update"
						namespace_id = scaleway_function_namespace.main.id
						runtime = "node14"
						privacy = "private"
						handler = "handler.handle"
					}

					resource scaleway_function_cron main {
						name = "name-changed"
						function_id = scaleway_function.main.id
						schedule = "0 0 * * *"
						args = jsonencode({test = "scw"})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionCronExists(tt, "scaleway_function_cron.main"),
					resource.TestCheckResourceAttr("scaleway_function_cron.main", "name", "name-changed"),
				),
			},
		},
	})
}

func TestAccScalewayFunctionCron_WithArgs(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionCronDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "tf-tests-function-cron-with-args"
					}

					resource scaleway_function main {
						name = "tf-tests-cron-with-args"
						namespace_id = scaleway_function_namespace.main.id
						runtime = "node14"
						privacy = "private"
						handler = "handler.handle"
					}

					resource scaleway_function_cron main {
						name = "tf-tests-cron-with-args"
						function_id = scaleway_function.main.id
						schedule = "0 0 * * *"
						args = jsonencode({test = "scw"})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionCronExists(tt, "scaleway_function_cron.main"),
					resource.TestCheckResourceAttr("scaleway_function_cron.main", "schedule", "0 0 * * *"),
					resource.TestCheckResourceAttr("scaleway_function_cron.main", "args", "{\"test\":\"scw\"}"),
					resource.TestCheckResourceAttr("scaleway_function_cron.main", "name", "tf-tests-cron-with-args"),
				),
			},
		},
	})
}

func testAccCheckScalewayFunctionCronExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := scaleway.FunctionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetCron(&function.GetCronRequest{
			CronID: id,
			Region: region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayFunctionCronDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_function_cron" {
				continue
			}

			api, region, id, err := scaleway.FunctionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteCron(&function.DeleteCronRequest{
				CronID: id,
				Region: region,
			})

			if err == nil {
				return fmt.Errorf("function cron (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
