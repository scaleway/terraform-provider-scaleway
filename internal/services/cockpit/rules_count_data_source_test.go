package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceCockpitRulesCount_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_rules_count_basic"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%s"
					}

					data "scaleway_cockpit_rules_count" "main" {
						project_id = scaleway_account_project.project.id
						region     = "fr-par"
					}
				`, projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_rules_count.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_rules_count.main", "region", "fr-par"),
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_rules_count.main", "project_id", "scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_rules_count.main", "preconfigured_rules_count"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_rules_count.main", "custom_rules_count"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_rules_count.main", "rules_count_by_datasource.#"),
				),
			},
		},
	})
}
