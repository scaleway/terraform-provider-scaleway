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

func TestAccActionRDBDatabaseBackupExport_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRDBDatabaseBackupExport_Basic because action are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_rdb_instance" "main" {
						name = "test-rdb-action-backup-export"
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
						name = "test-backup-export"
						depends_on = [scaleway_rdb_database.main]

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_rdb_database_export_backup.main]
							}
						}
					}

					action "scaleway_rdb_database_export_backup" "main" {
						config {
							backup_id = scaleway_rdb_database_backup.main.id
							wait = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isBackupExported(tt, "scaleway_rdb_database_backup.main"),
				),
			},
		},
	})
}

func isBackupExported(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		region, id, err := regional.ParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse backup ID: %w", err)
		}

		api := rdbSDK.NewAPI(tt.Meta.ScwClient())

		backup, err := api.GetDatabaseBackup(&rdbSDK.GetDatabaseBackupRequest{
			Region:           region,
			DatabaseBackupID: id,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to get backup: %w", err)
		}

		if backup == nil {
			return fmt.Errorf("backup %s not found", id)
		}

		if backup.Status != rdbSDK.DatabaseBackupStatusReady {
			return fmt.Errorf("backup %s is not ready, status: %s", id, backup.Status)
		}

		return nil
	}
}
