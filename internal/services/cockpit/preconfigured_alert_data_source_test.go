package cockpit_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceCockpitPreconfiguredAlert_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_preconfigured_alert_ds"
					}

					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
					}

					data "scaleway_cockpit_preconfigured_alert" "main" {
						project_id = scaleway_cockpit.main.project_id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_preconfigured_alert.main", "project_id", "scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_preconfigured_alert.main", "alerts.#"),
				),
			},
		},
	})
}

func TestAccDataSourceCockpitPreconfiguredAlert_WithFilters(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_preconfigured_alert_filters"
					}

					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
					}

					data "scaleway_cockpit_preconfigured_alert" "enabled" {
						project_id  = scaleway_cockpit.main.project_id
						rule_status = "enabled"
					}

					data "scaleway_cockpit_preconfigured_alert" "disabled" {
						project_id  = scaleway_cockpit.main.project_id
						rule_status = "disabled"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_preconfigured_alert.enabled", "alerts.#"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_preconfigured_alert.disabled", "alerts.#"),
				),
			},
		},
	})
}
