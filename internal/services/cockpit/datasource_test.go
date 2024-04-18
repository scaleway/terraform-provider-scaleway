package cockpit_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	cockpitSDK "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/cockpit"
)

func TestAccCockpitDatasource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isDatasourceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_project" "project" {
						name = "tf_tests_cockpit_datasource_basic"
				  	}

					resource "scaleway_cockpit_datasource" "main" {
					  project_id = scaleway_account_project.project.id
					  name       = "my-datasource"
					  type       = "metrics"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatasourcePresent(tt, "scaleway_cockpit_datasource.main"),
					resource.TestCheckResourceAttr("scaleway_cockpit_datasource.main", "name", "my-datasource"),
					resource.TestCheckResourceAttr("scaleway_cockpit_datasource.main", "type", "metrics"),
					resource.TestCheckResourceAttr("scaleway_cockpit_datasource.main", "region", "fr-par"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_datasource.main", "url"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_datasource.main", "origin"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_datasource.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_datasource.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_datasource.main", "synchronized_with_grafana"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit_datasource.main", "project_id", "scaleway_account_project.project", "id"),
				),
			},
		},
	})
}

func isDatasourcePresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource cockpit datasource not found: %s", n)
		}

		api, region, ID, err := cockpit.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDataSource(&cockpitSDK.RegionalAPIGetDataSourceRequest{
			Region:       region,
			DataSourceID: ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isDatasourceDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_cockpit_datasource" {
				continue
			}

			api, region, ID, err := cockpit.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetDataSource(&cockpitSDK.RegionalAPIGetDataSourceRequest{
				Region:       region,
				DataSourceID: ID,
			})

			if err == nil {
				return fmt.Errorf("cockpit datasource (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
