package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceRdbDatabase_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_rdb_instance" "server" {
						name      = "test-terraform"
						node_type = "db-dev-s"
						engine    = "PostgreSQL-11"
					}
					resource "scaleway_rdb_database" "database" {
						name        = "test-terraform"
						instance_id = scaleway_rdb_instance.server.id
					}`,
			},
			{
				Config: `
					resource "scaleway_rdb_instance" "server" {
						name      = "test-terraform"
						node_type = "db-dev-s"
						engine    = "PostgreSQL-11"
					}
					resource "scaleway_rdb_database" "database" {
						name        = "test-terraform"
						instance_id = scaleway_rdb_instance.server.id
					}
					data "scaleway_rdb_database" "find_by_name_and_instance" {
						name        = scaleway_rdb_database.database.name
						instance_id = scaleway_rdb_instance.server.id
					}
					data "scaleway_rdb_database" "find_by_instance" {
						instance_id = scaleway_rdb_instance.server.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbDatabaseExists(tt, "scaleway_rdb_instance.server", "scaleway_rdb_database.database"),

					resource.TestCheckResourceAttr("data.scaleway_rdb_database.find_by_name_and_instance", "name", "test-terraform"),
					resource.TestCheckResourceAttr("data.scaleway_rdb_database.find_by_name_and_instance", "managed", "true"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database.find_by_name_and_instance", "owner"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database.find_by_name_and_instance", "size"),

					resource.TestCheckResourceAttr("data.scaleway_rdb_database.find_by_instance", "name", "rdb"),
					resource.TestCheckResourceAttr("data.scaleway_rdb_database.find_by_instance", "managed", "true"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database.find_by_instance", "owner"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database.find_by_instance", "size"),
				),
			},
		},
	})
}
