package scaleway

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	accountV3 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_cockpit", &resource.Sweeper{
		Name: "scaleway_cockpit",
		F:    testSweepCockpit,
	})
}

func testSweepCockpit(_ string) error {
	return sweep(func(scwClient *scw.Client) error {
		accountAPI := accountV3.NewProjectAPI(scwClient)
		cockpitAPI := cockpit.NewAPI(scwClient)

		listProjects, err := accountAPI.ListProjects(&accountV3.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			_, err = cockpitAPI.WaitForCockpit(&cockpit.WaitForCockpitRequest{
				ProjectID: project.ID,
				Timeout:   scw.TimeDurationPtr(defaultCockpitTimeout),
			})
			if err != nil {
				if !is404Error(err) {
					return fmt.Errorf("failed to deactivate cockpit: %w", err)
				}
			}

			_, err = cockpitAPI.DeactivateCockpit(&cockpit.DeactivateCockpitRequest{
				ProjectID: project.ID,
			})
			if err != nil {
				if !is404Error(err) {
					return fmt.Errorf("failed to deactivate cockpit: %w", err)
				}
			}
		}

		return nil
	})
}

func TestAccScalewayCockpit_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayCockpitDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_project_basic"
				  	}

					resource scaleway_cockpit main {
						project_id = scaleway_account_project.project.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitExists(tt, "scaleway_cockpit.main"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "plan_id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.metrics_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.logs_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.alertmanager_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.grafana_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.traces_url"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit.main", "project_id", "scaleway_account_project.project", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_project_basic"
				  	}

					data "scaleway_cockpit_plan" "premium" {
						name = "premium"
					}

					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
						plan       = data.scaleway_cockpit_plan.premium.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitExists(tt, "scaleway_cockpit.main"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "plan_id"),
				),
			},
		},
	})
}

func TestAccScalewayCockpit_PremiumPlanByID(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayCockpitDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_project_premium"
				  	}

					data "scaleway_cockpit_plan" "premium" {
						name = "premium"
					}

					resource scaleway_cockpit main {
						project_id = scaleway_account_project.project.id
						plan       = data.scaleway_cockpit_plan.premium.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitExists(tt, "scaleway_cockpit.main"),
				),
			},
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_project_premium"
				  	}

					data "scaleway_cockpit_plan" "free" {
						name = "free"
					}

					resource scaleway_cockpit main {
						project_id = scaleway_account_project.project.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitExists(tt, "scaleway_cockpit.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit.main", "plan_id", "data.scaleway_cockpit_plan.free", "id"),
				),
			},
		},
	})
}

func TestAccScalewayCockpit_PremiumPlanByName(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayCockpitDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_project_premium"
				  	}

					data "scaleway_cockpit_plan" "premium" {
						name = "premium"
					}

					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
						plan       = "premium"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitExists(tt, "scaleway_cockpit.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit.main", "plan_id", "data.scaleway_cockpit_plan.premium", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayCockpitExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource cockpit not found: %s", n)
		}

		api, err := cockpitAPI(tt.Meta)
		if err != nil {
			return err
		}

		_, err = api.GetCockpit(&cockpit.GetCockpitRequest{
			ProjectID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayCockpitDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_cockpit" {
				continue
			}

			api, err := cockpitAPI(tt.Meta)
			if err != nil {
				return err
			}

			_, err = api.DeactivateCockpit(&cockpit.DeactivateCockpitRequest{
				ProjectID: rs.Primary.ID,
			})
			if err == nil {
				return fmt.Errorf("cockpit (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
