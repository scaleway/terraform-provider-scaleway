package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
)

func TestAccScalewayMNQNatsCredentials_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQNatsCredentialsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_mnq_nats_account main {
						name = "test-mnq-nats-credentials-basic"
					}

					resource scaleway_mnq_nats_credentials main {
						account_id = scaleway_mnq_nats_account.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQNatsCredentialsExists(tt, "scaleway_mnq_nats_credentials.main"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_nats_credentials.main", "file"),
				),
			},
		},
	})
}

func testAccCheckScalewayMNQNatsCredentialsExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := mnqNatsAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetNatsCredentials(&mnq.NatsAPIGetNatsCredentialsRequest{
			NatsCredentialsID: id,
			Region:            region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayMNQNatsCredentialsDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_nats_credentials" {
				continue
			}

			api, region, id, err := mnqNatsAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteNatsCredentials(&mnq.NatsAPIDeleteNatsCredentialsRequest{
				NatsCredentialsID: id,
				Region:            region,
			})

			if err == nil {
				return fmt.Errorf("mnq nats credentials (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
