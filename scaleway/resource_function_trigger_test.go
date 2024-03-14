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
	resource.AddTestSweepers("scaleway_function_trigger", &resource.Sweeper{
		Name: "scaleway_function_trigger",
		F:    testSweepFunctionTrigger,
	})
}

func testSweepFunctionTrigger(_ string) error {
	return sweepRegions((&function.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		functionAPI := function.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the function triggers in (%s)", region)
		listTriggers, err := functionAPI.ListTriggers(
			&function.ListTriggersRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing trigger in (%s) in sweeper: %s", region, err)
		}

		for _, trigger := range listTriggers.Triggers {
			_, err := functionAPI.DeleteTrigger(&function.DeleteTriggerRequest{
				TriggerID: trigger.ID,
				Region:    region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting trigger in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayFunctionTrigger_SQS(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	config := `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_function_trigger_sqs"
					}

					resource scaleway_function_namespace main {
						name = "test-function-trigger-sqs"	
						project_id = scaleway_account_project.project.id
					}

					resource scaleway_function main {
						name = "test-function-trigger-sqs"
						namespace_id = scaleway_function_namespace.main.id
						runtime = "node20"
						privacy = "private"
						handler = "handler.handle"
					}

					resource scaleway_mnq_sqs main {
						project_id = scaleway_account_project.project.id
					}

					resource "scaleway_mnq_sqs_credentials" "main" {
						project_id = scaleway_mnq_sqs.main.project_id
					
						permissions {
							can_publish = true
							can_receive = true
							can_manage  = true
						}
					}
					
					resource "scaleway_mnq_sqs_queue" "queue" {
						project_id = scaleway_mnq_sqs.main.project_id
						name = "TestQueue"
						access_key = scaleway_mnq_sqs_credentials.main.access_key
						secret_key = scaleway_mnq_sqs_credentials.main.secret_key
					}

					resource scaleway_function_trigger main {
						function_id = scaleway_function.main.id
						name = "test-function-trigger-sqs"
						sqs {
							queue = "TestQueue"
							project_id = scaleway_mnq_sqs.main.project_id
							region = scaleway_mnq_sqs.main.region
						}
					}
				`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionTriggerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionTriggerExists(tt, "scaleway_function_trigger.main"),
					testCheckResourceAttrUUID("scaleway_function_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_function_trigger.main", "name", "test-function-trigger-sqs"),
					testAccCheckScalewayFunctionTriggerStatusReady(tt, "scaleway_function_trigger.main"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func TestAccScalewayFunctionTrigger_Nats(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	config := `
					resource scaleway_function_namespace main {
						name = "test-function-trigger-sqs"	
					}

					resource scaleway_function main {
						name = "test-function-trigger-sqs"
						namespace_id = scaleway_function_namespace.main.id
						runtime = "node20"
						privacy = "private"
						handler = "handler.handle"
					}

					resource "scaleway_mnq_nats_account" "main" {}

					resource scaleway_function_trigger main {
						function_id = scaleway_function.main.id
						name = "test-function-trigger-nats"
						nats {
							subject = "TestSubject"
							account_id = scaleway_mnq_nats_account.main.id
							region = scaleway_mnq_nats_account.main.region
						}
					}
				`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionTriggerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionTriggerExists(tt, "scaleway_function_trigger.main"),
					testCheckResourceAttrUUID("scaleway_function_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_function_trigger.main", "name", "test-function-trigger-nats"),
					testAccCheckScalewayFunctionTriggerStatusReady(tt, "scaleway_function_trigger.main"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func TestAccScalewayFunctionTrigger_Error(t *testing.T) {
	// https://github.com/hashicorp/terraform-plugin-testing/issues/69
	t.Skip("Currently cannot test warnings")
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionTriggerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "test-function-trigger-error"	
					}

					resource scaleway_function main {
						name = "test-function-trigger-error"
						namespace_id = scaleway_function_namespace.main.id
						runtime = "node14"
						privacy = "private"
						handler = "handler.handle"
					}

					resource scaleway_mnq_namespace main {
						protocol = "sqs_sns"
						name = "test-function-trigger-error"
					}

					resource scaleway_function_trigger main {
						function_id = scaleway_function.main.id
						name = "test-function-trigger-error"
						sqs {
							namespace_id = scaleway_mnq_namespace.main.id
							queue = "TestQueue"
							project_id = scaleway_mnq_namespace.main.project_id
							region = scaleway_mnq_namespace.main.region
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionTriggerExists(tt, "scaleway_function_trigger.main"),
					testCheckResourceAttrUUID("scaleway_function_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_function_trigger.main", "name", "test-function-trigger-error"),
				),
			},
		},
	})
}

func testAccCheckScalewayFunctionTriggerExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := scaleway.FunctionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetTrigger(&function.GetTriggerRequest{
			TriggerID: id,
			Region:    region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayFunctionTriggerStatusReady(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := scaleway.FunctionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		trigger, err := api.GetTrigger(&function.GetTriggerRequest{
			TriggerID: id,
			Region:    region,
		})
		if err != nil {
			return err
		}

		if trigger.Status != function.TriggerStatusReady {
			return fmt.Errorf("trigger status is %s, expected ready", trigger.Status)
		}

		return nil
	}
}

func testAccCheckScalewayFunctionTriggerDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_function_trigger" {
				continue
			}

			api, region, id, err := scaleway.FunctionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteTrigger(&function.DeleteTriggerRequest{
				TriggerID: id,
				Region:    region,
			})

			if err == nil {
				return fmt.Errorf("function trigger (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
