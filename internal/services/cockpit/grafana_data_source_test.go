package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceCockpitGrafana_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_data_basic"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%s"
					}

					resource "scaleway_cockpit" "main" {
						project_id = scaleway_account_project.project.id
					}

					data "scaleway_cockpit_grafana" "main" {
						project_id = scaleway_cockpit.main.project_id
					}
				`, projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_grafana.main", "project_id", "scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana.main", "grafana_url"),
				),
			},
		},
	})
}

func TestAccDataSourceCockpitGrafana_DefaultProject(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	config := `
		data scaleway_cockpit_grafana main {
		}
	`

	_, projectIDExists := tt.Meta.ScwClient().GetDefaultProjectID()
	if !projectIDExists {
		config = `
			data scaleway_cockpit_grafana main {
				project_id = "105bdce1-64c0-48ab-899d-868455867ecf"
			}
		`
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana.main", "project_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_grafana.main", "grafana_url"),
				),
			},
		},
	})
}
