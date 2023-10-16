package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
	func init() {
		resource.AddTestSweepers("scaleway_nats_queue", &resource.Sweeper{
			Name: "scaleway_nats_queue",
			F:    testSweepNATSQueue,
		})
	}

	func testSweepNATSQueue(_ string) error {
		return sweepRegions((&mnq.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
			mnqAPI := mnq.NewAPI(scwClient)
			l.Debugf("sweeper: destroying the mnq natsqueues in (%s)", region)
			listNATSQueues, err := mnqAPI.ListNATSQueues(
				&mnq.ListNATSQueuesRequest{
					Region: region,
				}, scw.WithAllPages())
			if err != nil {
				return fmt.Errorf("error listing natsqueue in (%s) in sweeper: %s", region, err)
			}

			for _, natsqueue := range listNATSQueues.NATSQueues {
				_, err := mnqAPI.DeleteNATSQueue(&mnq.DeleteNATSQueueRequest{
					NATSQueueID: natsqueue.ID,
					Region:      region,
				})
				if err != nil {
					l.Debugf("sweeper: error (%s)", err)

					return fmt.Errorf("error deleting natsqueue in sweeper: %s", err)
				}
			}

			return nil
		})
	}
*/
func TestAccScalewayMNQNatsQueue_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayNatsQueueDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_mnq_nats_account main {
						name = "test-mnq-nats-queue-basic"
					}

					resource scaleway_mnq_nats_credentials main {
						account_id = scaleway_mnq_nats_account.main.id
					}

					resource scaleway_mnq_nats_queue main {
						name = "test-mnq-nats-queue-basic"
						endpoint = scaleway_mnq_nats_account.main.endpoint
						credentials = scaleway_mnq_nats_credentials.main.file
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayNatsQueueExists(tt, "scaleway_mnq_nats_queue.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_nats_queue.main", "name", "test-mnq-nats-queue-basic"),
				),
			},
		},
	})
}

func testAccCheckScalewayNatsQueueExists(_ *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		region, _, queueName, err := decomposeMNQQueueID(rs.Primary.ID)
		if err != nil {
			return err
		}

		js, err := newNATSJetStreamClient(region.String(), rs.Primary.Attributes["endpoint"], rs.Primary.Attributes["credentials"])
		if err != nil {
			return err
		}

		_, err = js.StreamInfo(queueName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayNatsQueueDestroy(_ *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_nats_queue" {
				continue
			}

			region, _, queueName, err := decomposeMNQQueueID(rs.Primary.ID)
			if err != nil {
				return err
			}

			js, err := newNATSJetStreamClient(region.String(), rs.Primary.Attributes["endpoint"], rs.Primary.Attributes["credentials"])
			if err != nil {
				return err
			}

			_, err = js.StreamInfo(queueName)

			if err == nil {
				return fmt.Errorf("mnq natsqueue (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}
