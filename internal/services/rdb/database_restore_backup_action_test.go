package rdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func TestAccActionRDBDatabaseBackupRestore_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRDBDatabaseBackupRestore_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_rdb_instance" "main" {
						name = "test-rdb-action-backup-restore"
						node_type = "db-dev-s"
						engine = "PostgreSQL-15"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						volume_type = "sbs_5k"
						volume_size_in_gb = 10
					}

					resource "scaleway_rdb_database" "main" {
						instance_id = scaleway_rdb_instance.main.id
						name = "test_db"
					}

					resource "scaleway_rdb_database_backup" "main" {
						instance_id = scaleway_rdb_instance.main.id
						database_name = scaleway_rdb_database.main.name
						name = "test-backup"
						depends_on = [scaleway_rdb_database.main]
					}

					resource "scaleway_rdb_instance" "restore_target" {
						name = "test-rdb-action-backup-restore-target"
						node_type = "db-dev-s"
						engine = "PostgreSQL-15"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						volume_type = "sbs_5k"
						volume_size_in_gb = 10

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_rdb_database_restore_backup.main]
							}
						}
					}

					action "scaleway_rdb_database_restore_backup" "main" {
						config {
							backup_id = scaleway_rdb_database_backup.main.id
							instance_id = scaleway_rdb_instance.restore_target.id
							database_name = "test_db"
							wait = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDatabaseRestored(tt, "scaleway_rdb_instance.restore_target", "test_db"),
				),
			},
		},
	})
}

func isDatabaseRestored(tt *acctest.TestTools, instanceResourceName, databaseName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[instanceResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", instanceResourceName)
		}

		region, id, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse instance ID: %w", err)
		}

		api := rdbSDK.NewAPI(tt.Meta.ScwClient())

		databases, err := api.ListDatabases(&rdbSDK.ListDatabasesRequest{
			Region:     region,
			InstanceID: id,
			Name:       new(databaseName),
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to list databases: %w", err)
		}

		if len(databases.Databases) == 0 {
			return fmt.Errorf("database %s not found in instance %s", databaseName, id)
		}

		return nil
	}
}
