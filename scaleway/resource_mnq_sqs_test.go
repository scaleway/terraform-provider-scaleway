package scaleway

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	accountV3 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func init() {
	resource.AddTestSweepers("scaleway_mnq_sqs", &resource.Sweeper{
		Name: "scaleway_mnq_sqs",
		F:    testSweepMNQSQS,
	})
}

func testSweepMNQSQS(_ string) error {
	return sweepRegions((&mnq.SqsAPI{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		accountAPI := accountV3.NewProjectAPI(scwClient)
		mnqAPI := mnq.NewSqsAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the mnq sqss in (%s)", region)

		listProjects, err := accountAPI.ListProjects(&accountV3.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}
		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			_, err := mnqAPI.DeactivateSqs(&mnq.SqsAPIDeactivateSqsRequest{
				Region:    region,
				ProjectID: project.ID,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)
				return err
			}
		}

		return nil
	})
}

func TestAccScalewayMNQSQS_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQSQSDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_account_project main {
						name = "tf_tests_mnq_sqs_basic"
					}

					resource scaleway_mnq_sqs main {
						project_id = scaleway_account_project.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayMNQSQSExists(tt, "scaleway_mnq_sqs.main"),
					testCheckResourceAttrUUID("scaleway_mnq_sqs.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_mnq_sqs.main", "endpoint"),
				),
			},
		},
	})
}

func TestAccScalewayMNQSQS_AlreadyActivated(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayMNQSQSDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_account_project main {
						name = "tf_tests_mnq_sqs_already_activated"
					}

					resource scaleway_mnq_sqs main {
						project_id = scaleway_account_project.main.id
					}
				`,
			},
			{
				Config: `
					resource scaleway_account_project main {
						name = "tf_tests_mnq_sqs_already_activated"
					}

					resource scaleway_mnq_sqs main {
						project_id = scaleway_account_project.main.id
					}

					resource scaleway_mnq_sqs duplicated {
						project_id = scaleway_account_project.main.id
					}
				`,
				ExpectError: regexp.MustCompile(".*Conflict.*"),
			},
		},
	})
}

func testAccCheckScalewayMNQSQSExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := mnqSQSAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		sqs, err := api.GetSqsInfo(&mnq.SqsAPIGetSqsInfoRequest{
			ProjectID: id,
			Region:    region,
		})
		if err != nil {
			return err
		}

		if sqs.Status != mnq.SqsInfoStatusEnabled {
			return fmt.Errorf("sqs status should be enabled, got: %s", sqs.Status)
		}

		return nil
	}
}

func testAccCheckScalewayMNQSQSDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mnq_sqs" {
				continue
			}

			api, region, id, err := mnqSQSAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			sqs, err := api.DeactivateSqs(&mnq.SqsAPIDeactivateSqsRequest{
				ProjectID: id,
				Region:    region,
			})
			if err != nil {
				if is404Error(err) { // Project may have been deleted
					return nil
				}
				return err
			}

			if sqs.Status != mnq.SqsInfoStatusDisabled {
				return fmt.Errorf("mnq sqs (%s) should be disabled", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
