package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceCockpitGrafanaProductDashboard_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_product_dashboard"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%s"
					}

					data "scaleway_cockpit_grafana_product_dashboards" "all" {
						project_id = scaleway_account_project.project.id
					}

					data "scaleway_cockpit_grafana_product_dashboard" "main" {
						project_id      = scaleway_account_project.project.id
						dashboard_name  = data.scaleway_cockpit_grafana_product_dashboards.all.names[0]
					}
				`, projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_grafana_product_dashboard.main", "project_id", "scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboard.main", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboard.main", "name"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboard.main", "title"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboard.main", "url"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_cockpit_grafana_product_dashboard.main", "name",
						"data.scaleway_cockpit_grafana_product_dashboards.all", "names.0",
					),
				),
			},
		},
	})
}
