package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_container_trigger", &resource.Sweeper{
		Name: "scaleway_container_trigger",
		F:    testSweepContainerTrigger,
	})
}

func testSweepContainerTrigger(_ string) error {
	return sweepRegions((&container.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		containerAPI := container.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the container triggers in (%s)", region)
		listTriggers, err := containerAPI.ListTriggers(
			&container.ListTriggersRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing trigger in (%s) in sweeper: %s", region, err)
		}

		for _, trigger := range listTriggers.Triggers {
			_, err := containerAPI.DeleteTrigger(&container.DeleteTriggerRequest{
				TriggerID: trigger.ID,
				Region:    region,
			})
			if err != nil {
				l.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting trigger in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayContainerTrigger_SQS(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	basicConfig := `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
					}

					resource scaleway_mnq_namespace main {
						protocol = "sqs_sns"
						name = "test-container-trigger-sqs"
					}

					resource "scaleway_mnq_credential" "main" {
					  namespace_id = scaleway_mnq_namespace.main.id
					
					  sqs_sns_credentials {
						permissions {
						  can_publish = true
						  can_receive = true
						  can_manage  = true
						}
					  }
					}
					
					resource "scaleway_mnq_queue" "queue" {
					  namespace_id = scaleway_mnq_namespace.main.id
					  name = "TestQueue"
					  sqs {
						access_key = scaleway_mnq_credential.main.sqs_sns_credentials.0.access_key
						secret_key = scaleway_mnq_credential.main.sqs_sns_credentials.0.secret_key
					  }
					}

					resource scaleway_container_trigger main {
						container_id = scaleway_container.main.id
						name = "test-container-trigger-sqs"
						sqs {
							namespace_id = scaleway_mnq_namespace.main.id
							queue = "TestQueue"
							project_id = scaleway_mnq_namespace.main.project_id
							region = scaleway_mnq_namespace.main.region
						}
					}
				`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerTriggerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: basicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerTriggerExists(tt, "scaleway_container_trigger.main"),
					testCheckResourceAttrUUID("scaleway_container_trigger.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container_trigger.main", "name", "test-container-trigger-sqs"),
					testAccCheckScalewayContainerTriggerStatusReady(tt, "scaleway_container_trigger.main"),
				),
			},
			{
				Config:   basicConfig,
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckScalewayContainerTriggerExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetTrigger(&container.GetTriggerRequest{
			TriggerID: id,
			Region:    region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayContainerTriggerStatusReady(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		trigger, err := api.GetTrigger(&container.GetTriggerRequest{
			TriggerID: id,
			Region:    region,
		})
		if err != nil {
			return err
		}

		if trigger.Status != container.TriggerStatusReady {
			return fmt.Errorf("trigger status is %s, expected ready", trigger.Status)
		}

		return nil
	}
}

func testAccCheckScalewayContainerTriggerDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_container_trigger" {
				continue
			}

			api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteTrigger(&container.DeleteTriggerRequest{
				TriggerID: id,
				Region:    region,
			})

			if err == nil {
				return fmt.Errorf("container trigger (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
