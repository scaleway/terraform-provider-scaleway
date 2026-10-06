package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceCockpitGrafanaProductDashboards_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_product_dashboards"

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
				`, projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_grafana_product_dashboards.all", "project_id", "scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboards.all", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboards.all", "dashboards.#"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboards.all", "names.#"),
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_grafana_product_dashboards.all", "dashboards.#", "data.scaleway_cockpit_grafana_product_dashboards.all", "names.#"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboards.all", "dashboards.0.name"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboards.all", "dashboards.0.title"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboards.all", "dashboards.0.url"),
				),
			},
		},
	})
}

func TestAccDataSourceCockpitGrafanaProductDashboards_WithTags(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_product_dashboards_tags"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%s"
					}

					data "scaleway_cockpit_grafana_product_dashboards" "filtered" {
						project_id = scaleway_account_project.project.id
						tags       = ["rdb"]
					}
				`, projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_grafana_product_dashboards.filtered", "project_id", "scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana_product_dashboards.filtered", "dashboards.#"),
				),
			},
		},
	})
}
