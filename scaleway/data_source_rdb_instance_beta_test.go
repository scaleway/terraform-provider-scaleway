package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewayDataSourceRDBInstance_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayRdbInstanceBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_rdb_instance_beta" "test" {
						name = "test-terraform"
						engine = "PostgreSQL-11"
						node_type = "db-dev-s"
					}

					data "scaleway_rdb_instance_beta" "test" {
						name = scaleway_rdb_instance_beta.test.name
					}

					data "scaleway_rdb_instance_beta" "test2" {
						instance_id = scaleway_rdb_instance_beta.test.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRdbBetaExists("scaleway_rdb_instance_beta.test"),

					resource.TestCheckResourceAttr("scaleway_rdb_instance_beta.test", "name", "test-terraform"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_beta.test", "id"),

					resource.TestCheckResourceAttr("data.scaleway_rdb_instance_beta.test2", "name", "test-terraform"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_instance_beta.test2", "id"),
				),
			},
		},
	})
}
