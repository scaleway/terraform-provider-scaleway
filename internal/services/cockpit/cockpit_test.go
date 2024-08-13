package cockpit_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
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
						plan       = "free"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("scaleway_cockpit.main", "project_id", "scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "plan"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "plan_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit.main", "plan", "free"),
				),
			},
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_project_basic"
				  	}
					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
						plan       = "premium"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "plan"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "plan_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit.main", "plan", "premium"),
				),
			},
		},
	})
}

func TestAccCockpit_WithSourceEndpoints(t *testing.T) {
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
					
					resource "scaleway_cockpit_source" "metrics" {
						project_id = scaleway_account_project.project.id
						name       = "my-data-source-metrics"
						type       = "metrics"
					}
					
					resource "scaleway_cockpit_source" "logs" {
						project_id = scaleway_account_project.project.id
						name       = "my-data-source-logs"
						type       = "logs"
					}
					
					resource "scaleway_cockpit_source" "traces" {
						project_id = scaleway_account_project.project.id
						name       = "my-data-source-traces"
						type       = "traces"
					}
					
					resource "scaleway_cockpit_alert_manager" "alert_manager" {
						project_id = scaleway_account_project.project.id
						enable_managed_alerts = true
					}

					resource "scaleway_cockpit_grafana_user" "main" {
					  project_id = scaleway_account_project.project.id
					  login = "cockpit_test_endpoint"
					  role = "editor"
					}
					
					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
						plan       = "premium"
						depends_on = [
								scaleway_cockpit_source.metrics,
								scaleway_cockpit_source.logs,
								scaleway_cockpit_source.traces,
								scaleway_cockpit_alert_manager.alert_manager,
								scaleway_cockpit_grafana_user.main
							]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_cockpit.main", "plan", "premium"),
					resource.TestCheckResourceAttr("scaleway_cockpit.main", "plan_id", "premium"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.metrics_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.logs_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.alertmanager_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.grafana_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "endpoints.0.traces_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "push_url.0.push_logs_url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit.main", "push_url.0.push_metrics_url"),
				),
			},
		},
	})
}

func isCockpitDestroyed(_ *acctest.TestTools) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		return nil
	}
}
