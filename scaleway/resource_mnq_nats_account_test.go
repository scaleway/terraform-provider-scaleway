package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/errs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func init() {
	resource.AddTestSweepers("scaleway_mnq_nats_account", &resource.Sweeper{
		Name: "scaleway_mnq_nats_account",
		F:    testSweepMNQNatsAccount,
	})
}

func testSweepMNQNatsAccount(_ string) error {
	return sweepRegions((&mnq.NatsAPI{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		mnqAPI := mnq.NewNatsAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the mnq nats accounts in (%s)", region)
		listNatsAccounts, err := mnqAPI.ListNatsAccounts(
			&mnq.NatsAPIListNatsAccountsRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing nats account in (%s) in sweeper: %s", region, err)
		}

		for _, account := range listNatsAccounts.NatsAccounts {
			err := mnqAPI.DeleteNatsAccount(&mnq.NatsAPIDeleteNatsAccountRequest{
				NatsAccountID: account.ID,
				Region:        region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting nats account in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayMNQNatsAccount_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQNatsAccountDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_mnq_nats_account main {
						name = "test-mnq-nats-account-basic"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQNatsAccountExists(tt, "scaleway_mnq_nats_account.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_nats_account.main", "name", "test-mnq-nats-account-basic"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_nats_account.main", "endpoint"),
				),
			},
		},
	})
}

func testAccCheckScalewayMNQNatsAccountExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := mnqNatsAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetNatsAccount(&mnq.NatsAPIGetNatsAccountRequest{
			NatsAccountID: id,
			Region:        region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayMNQNatsAccountDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_nats_account" {
				continue
			}

			api, region, id, err := mnqNatsAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteNatsAccount(&mnq.NatsAPIDeleteNatsAccountRequest{
				NatsAccountID: id,
				Region:        region,
			})

			if err == nil {
				return fmt.Errorf("mnq nats account (%s) still exists", rs.Primary.ID)
			}

			if !errs.Is404Error(err) {
				return err
			}
		}

		return nil
	}
}
