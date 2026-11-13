package jobs_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	jobsSDK "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha2"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/jobs"
)

const jobDefinitionRepeatedConfig = `
resource scaleway_job_definition main {
	name = "test-jobs-job-trigger-basic"
	cpu_limit = 120
	memory_limit = 256
	local_storage_capacity = 5120
	image_uri = "docker.io/alpine:latest"
}
`

func TestAccJobTrigger_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckJobTriggerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: jobDefinitionRepeatedConfig + `
					resource scaleway_job_trigger main {
						job_definition_id = scaleway_job_definition.main.id
						cron_config = {
							schedule = "12 13 * * *"
							timezone = "Europe/Paris"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					testAccCheckJobTriggerExists(tt, "scaleway_job_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_trigger.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_job_trigger.main", "name"),
					resource.TestCheckResourceAttrPair("scaleway_job_trigger.main", "job_definition_id",
						"scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_job_trigger.main", "cron_config.%"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.schedule", "12 13 * * *"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.timezone", "Europe/Paris"),
				),
			},
			{
				Config: jobDefinitionRepeatedConfig + `
					resource scaleway_job_trigger main {
						name = "test-jobs-job-trigger-basic"
						job_definition_id = scaleway_job_definition.main.id
						cron_config = {
							schedule        = "0 * 13 12 *"
							timezone        = "Europe/Paris"
							startup_command = [ "echo" ]
							args            = [ "test", "ok" ]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					testAccCheckJobTriggerExists(tt, "scaleway_job_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "name", "test-jobs-job-trigger-basic"),
					resource.TestCheckResourceAttrPair("scaleway_job_trigger.main", "job_definition_id",
						"scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_job_trigger.main", "cron_config.%"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.schedule", "0 * 13 12 *"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.timezone", "Europe/Paris"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.startup_command.#", "1"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.startup_command.0", "echo"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.args.#", "2"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.args.0", "test"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.args.1", "ok"),
				),
			},
			{
				Config: jobDefinitionRepeatedConfig + `
					resource scaleway_job_trigger main {
						name = "test-jobs-job-trigger-basic"
						job_definition_id = scaleway_job_definition.main.id
						cron_config = {
							schedule        = "12 21 13 12 *"
							timezone        = "Asia/Tokyo"
							startup_command = [ "pwd" ]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(tt, "scaleway_job_definition.main"),
					testAccCheckJobTriggerExists(tt, "scaleway_job_trigger.main"),
					acctest.CheckResourceAttrUUID("scaleway_job_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "name", "test-jobs-job-trigger-basic"),
					resource.TestCheckResourceAttrPair("scaleway_job_trigger.main", "job_definition_id",
						"scaleway_job_definition.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_job_trigger.main", "cron_config.%"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.schedule", "12 21 13 12 *"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.timezone", "Asia/Tokyo"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.startup_command.#", "1"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.startup_command.0", "pwd"),
					resource.TestCheckResourceAttr("scaleway_job_trigger.main", "cron_config.args.#", "0"),
				),
			},
			{
				ResourceName:      "scaleway_job_trigger.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckJobTriggerExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := jobs.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetTrigger(&jobsSDK.GetTriggerRequest{
			TriggerID: id,
			Region:    region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckJobTriggerDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_job_trigger" {
				continue
			}

			api, region, id, err := jobs.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteTrigger(&jobsSDK.DeleteTriggerRequest{
				TriggerID: id,
				Region:    region,
			})
			if err == nil {
				return fmt.Errorf("job trigger (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
