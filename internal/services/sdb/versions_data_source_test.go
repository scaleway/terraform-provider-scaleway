package sdb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccServerlessSQLDBVersionsDataSource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	versions := fetchAvailableVersions(tt)
	filterVersion := versions[0]

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_sdb_sql_versions" "pg" {}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_sdb_sql_versions.pg", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_sdb_sql_versions.pg", "region"),
					resource.TestCheckResourceAttrSet("data.scaleway_sdb_sql_versions.pg", "versions.0.name"),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_sdb_sql_versions" "pg" {
						name = %q
					}
				`, filterVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_sdb_sql_versions.pg", "id"),
					resource.TestCheckResourceAttr("data.scaleway_sdb_sql_versions.pg", "versions.0.name", filterVersion),
				),
			},
		},
	})
}
