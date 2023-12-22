package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_mnq_sns_credentials", &resource.Sweeper{
		Name: "scaleway_mnq_sns_credentials",
		F:    testSweepMNQSNSCredentials,
	})
}

func testSweepMNQSNSCredentials(_ string) error {
	return sweepRegions((&mnq.SnsAPI{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		mnqAPI := mnq.NewSnsAPI(scwClient)
		l.Debugf("sweeper: destroying the mnq sns credentials in (%s)", region)
		listSnsCredentials, err := mnqAPI.ListSnsCredentials(
			&mnq.SnsAPIListSnsCredentialsRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing sns credentials in (%s) in sweeper: %s", region, err)
		}

		for _, credentials := range listSnsCredentials.SnsCredentials {
			err := mnqAPI.DeleteSnsCredentials(&mnq.SnsAPIDeleteSnsCredentialsRequest{
				SnsCredentialsID: credentials.ID,
				Region:           region,
			})
			if err != nil {
				l.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting sns credentials in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayMNQSNSCredentials_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQSNSCredentialsDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_account_project main {}

					resource scaleway_mnq_sns main {
						project_id = scaleway_account_project.main.id
					}

					resource scaleway_mnq_sns_credentials main {
						project_id = scaleway_mnq_sns.main.project_id
						name = "test-mnq-sns-credentials-basic"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQSNSCredentialsExists(tt, "scaleway_mnq_sns_credentials.main"),
					testCheckResourceAttrUUID("scaleway_mnq_sns_credentials.main", "id"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_credentials.main", "name", "test-mnq-sns-credentials-basic"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_sns_credentials.main", "access_key"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_sns_credentials.main", "secret_key"),
				),
			},
			{
				Config: `
					resource scaleway_account_project main {}

					resource scaleway_mnq_sns main {
						project_id = scaleway_account_project.main.id
					}

					resource scaleway_mnq_sns_credentials main {
						project_id = scaleway_mnq_sns.main.project_id
						name = "test-mnq-sns-credentials-basic"
						permissions {
							can_manage = true
							can_receive = false
							can_publish = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQSNSCredentialsExists(tt, "scaleway_mnq_sns_credentials.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_credentials.main", "permissions.0.can_manage", "true"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_credentials.main", "permissions.0.can_receive", "false"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_credentials.main", "permissions.0.can_publish", "true"),
				),
			},
			{
				Config: `
					resource scaleway_account_project main {}

					resource scaleway_mnq_sns main {
						project_id = scaleway_account_project.main.id
					}

					resource scaleway_mnq_sns_credentials main {
						project_id = scaleway_mnq_sns.main.project_id
						name = "test-mnq-sns-credentials-basic"
						permissions {
							can_manage = false
							can_receive = true
							can_publish = false
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQSNSCredentialsExists(tt, "scaleway_mnq_sns_credentials.main"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_credentials.main", "permissions.0.can_manage", "false"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_credentials.main", "permissions.0.can_receive", "true"),
					resource.TestCheckResourceAttr("scaleway_mnq_sns_credentials.main", "permissions.0.can_publish", "false"),
				),
			},
		},
	})
}

func testAccCheckScalewayMNQSNSCredentialsExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := mnqSNSAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetSnsCredentials(&mnq.SnsAPIGetSnsCredentialsRequest{
			SnsCredentialsID: id,
			Region:           region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayMNQSNSCredentialsDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_sns_credentials" {
				continue
			}

			api, region, id, err := mnqSNSAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteSnsCredentials(&mnq.SnsAPIDeleteSnsCredentialsRequest{
				SnsCredentialsID: id,
				Region:           region,
			})

			if err == nil {
				return fmt.Errorf("mnq sns credentials (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
