package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceRdbDatabaseBackup_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbDatabaseBackupDestroy(tt),
		),
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

					resource scaleway_rdb_database_backup backup {
						instance_id 	= scaleway_rdb_instance.server.id
  						database_name 	= scaleway_rdb_database.database.name
  						name 			= "test_backup_datasource"
					}

					data scaleway_rdb_database_backup find_by_name {
						name        = scaleway_rdb_database_backup.backup.name
					}

					data scaleway_rdb_database_backup find_by_name_and_instance {
						name        = scaleway_rdb_database_backup.backup.name
						instance_id = scaleway_rdb_instance.server.id
					}

					data scaleway_rdb_database_backup find_by_id {
						backup_id 	= scaleway_rdb_database_backup.backup.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbDatabaseExists(tt, "scaleway_rdb_instance.server", "scaleway_rdb_database.database"),
					testAccCheckRdbDatabaseBackupExists(tt, "scaleway_rdb_database_backup.backup"),

					resource.TestCheckResourceAttr("data.scaleway_rdb_database_backup.find_by_name", "name", "test_backup_datasource"),
					resource.TestCheckResourceAttr("data.scaleway_rdb_database_backup.find_by_name_and_instance", "name", "test_backup_datasource"),
					resource.TestCheckResourceAttr("data.scaleway_rdb_database_backup.find_by_id", "name", "test_backup_datasource"),
				),
			},
		},
	})
}
