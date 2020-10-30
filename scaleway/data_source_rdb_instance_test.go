package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceRDBInstance_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_rdb_instance_beta" "test" {
						name = "test-terraform"
						engine = "PostgreSQL-11"
						node_type = "db-dev-s"
					}`,
			},
			{
				Config: `
					resource "scaleway_rdb_instance_beta" "test" {
						name = "test-terraform"
						engine = "PostgreSQL-11"
						node_type = "db-dev-s"
					}

					data "scaleway_rdb_instance" "test" {
						name = scaleway_rdb_instance_beta.test.name
					}

					data "scaleway_rdb_instance" "test2" {
						instance_id = scaleway_rdb_instance_beta.test.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRdbBetaExists(tt, "scaleway_rdb_instance_beta.test"),

					resource.TestCheckResourceAttr("scaleway_rdb_instance_beta.test", "name", "test-terraform"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance.test", "id"),

					resource.TestCheckResourceAttr("data.scaleway_rdb_instance.test2", "name", "test-terraform"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance.test2", "id"),
				),
			},
		},
	})
}
