package scaleway

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	accountV3 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_mnq_sns", &resource.Sweeper{
		Name: "scaleway_mnq_sns",
		F:    testSweepMNQSNS,
	})
}

func testSweepMNQSNS(_ string) error {
	return sweepRegions((&mnq.SnsAPI{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		accountAPI := accountV3.NewProjectAPI(scwClient)
		mnqAPI := mnq.NewSnsAPI(scwClient)

		l.Debugf("sweeper: destroying the mnq sns in (%s)", region)

		listProjects, err := accountAPI.ListProjects(&accountV3.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}
		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			_, err := mnqAPI.DeactivateSns(&mnq.SnsAPIDeactivateSnsRequest{
				Region:    region,
				ProjectID: project.ID,
			})
			if err != nil {
				l.Debugf("sweeper: error (%s)", err)
				return err
			}
		}

		return nil
	})
}

func TestAccScalewayMNQSNS_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQSNSTopicDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_account_project main {}

					resource scaleway_mnq_sns main {
						project_id = scaleway_account_project.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQSNSExists(tt, "scaleway_mnq_sns.main"),
					testCheckResourceAttrUUID("scaleway_mnq_sns.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_sns.main", "endpoint"),
				),
			},
		},
	})
}

func testAccCheckScalewayMNQSNSExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := mnqSNSAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		sns, err := api.GetSnsInfo(&mnq.SnsAPIGetSnsInfoRequest{
			ProjectID: id,
			Region:    region,
		})

		if sns.Status != mnq.SnsInfoStatusEnabled {
			return fmt.Errorf("sns status should be enabled, got: %s", sns.Status)
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayMNQSNSTopicDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_sns" {
				continue
			}

			api, region, id, err := mnqSNSAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			sns, err := api.DeactivateSns(&mnq.SnsAPIDeactivateSnsRequest{
				ProjectID: id,
				Region:    region,
			})
			if err != nil {
				if is404Error(err) {
					return nil
				}
				return err
			}

			if sns.Status != mnq.SnsInfoStatusDisabled {
				return fmt.Errorf("mnq sns (%s) should be disabled", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
