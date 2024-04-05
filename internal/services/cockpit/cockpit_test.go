package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	cockpitSDK "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/cockpit"
)

func TestAccCockpit_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isCockpitDestroyed(tt),
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
					isCockpitPresent(tt, "scaleway_cockpit.main"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "plan_id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.metrics_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.logs_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.alertmanager_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.grafana_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.traces_url"),
					resource.TestCheckResourceAttr("scaleway_cockpit.main", "push_url.0.push_logs_url", "https://logs.cockpit.fr-par.scw.cloud/loki/api/v1/push"),
					resource.TestCheckResourceAttr("scaleway_cockpit.main", "push_url.0.push_metrics_url", "https://metrics.cockpit.fr-par.scw.cloud/api/v1/push"),

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
					isCockpitPresent(tt, "scaleway_cockpit.main"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "plan_id"),
				),
			},
		},
	})
}

func TestAccCockpit_PremiumPlanByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isCockpitDestroyed(tt),
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
					isCockpitPresent(tt, "scaleway_cockpit.main"),
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
					isCockpitPresent(tt, "scaleway_cockpit.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit.main", "plan_id", "data.scaleway_cockpit_plan.free", "id"),
				),
			},
		},
	})
}

func TestAccCockpit_PremiumPlanByName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isCockpitDestroyed(tt),
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
					isCockpitPresent(tt, "scaleway_cockpit.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit.main", "plan_id", "data.scaleway_cockpit_plan.premium", "id"),
				),
			},
		},
	})
}

func isCockpitPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource cockpit not found: %s", n)
		}

		api, err := cockpit.NewAPI(tt.Meta)
		if err != nil {
			return err
		}

		_, err = api.GetCockpit(&cockpitSDK.GetCockpitRequest{
			ProjectID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isCockpitDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_cockpit" {
				continue
			}

			api, err := cockpit.NewAPI(tt.Meta)
			if err != nil {
				return err
			}

			_, err = api.DeactivateCockpit(&cockpitSDK.DeactivateCockpitRequest{
				ProjectID: rs.Primary.ID,
			})
			if err == nil {
				return fmt.Errorf("cockpit (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
