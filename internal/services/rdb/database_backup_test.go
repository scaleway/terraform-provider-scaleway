package rdb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb"
	rdbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb/testfuncs"
)

func init() {
	resource.AddTestSweepers("scaleway_rdb_database_backup", &resource.Sweeper{
		Name: "scaleway_rdb_database_backup",
		F:    testSweepDatabaseBackup,
	})
}

func testSweepDatabaseBackup(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		rdbAPI := rdbSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the rdb database backups in (%s)", region)

		listBackups, err := rdbAPI.ListDatabaseBackups(&rdbSDK.ListDatabaseBackupsRequest{
			Region: region,
		})
		if err != nil {
			return fmt.Errorf("error listing rdb database backups in (%s) in sweeper: %w", region, err)
		}

		for _, backup := range listBackups.DatabaseBackups {
			_, err := rdbAPI.DeleteDatabaseBackup(&rdbSDK.DeleteDatabaseBackupRequest{
				Region:           region,
				DatabaseBackupID: backup.ID,
			})
			if err != nil && !httperrors.Is404(err) {
				return fmt.Errorf("error deleting rdb database backup in sweeper: %w", err)
			}
		}

		return nil
	})
}

func TestAccDatabaseBackup_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	instanceName := "TestAccScalewayRdbDatabaseBackup_Basic"
	latestEngineVersion := rdbchecks.GetLatestEngineVersion(tt, postgreSQLEngineName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			rdbchecks.IsInstanceDestroyed(tt),
			isBackupDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance main {
						name = "%s"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
					}

					resource scaleway_rdb_database main {
						instance_id = scaleway_rdb_instance.main.id
						name = "foo"
					}`, instanceName, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					isDatabasePresent(tt, "scaleway_rdb_instance.main", "scaleway_rdb_database.main"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance main {
						name = "%s"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
					}

					resource scaleway_rdb_database main {
						instance_id = scaleway_rdb_instance.main.id
						name = "foo"
					}

					resource scaleway_rdb_database_backup main {
						instance_id = scaleway_rdb_instance.main.id
  						database_name = scaleway_rdb_database.main.name
  						name = "test_backup"
					}`, instanceName, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					isBackupPresent(tt, "scaleway_rdb_database_backup.main"),

					resource.TestCheckResourceAttr("scaleway_rdb_database_backup.main", "database_name", "foo"),
					resource.TestCheckResourceAttr("scaleway_rdb_database_backup.main", "name", "test_backup"),
				),
			},
		},
	})
}

func isBackupDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, rdbchecks.DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_rdb_database_backup" {
					continue
				}

				api, region, id, err := rdb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetDatabaseBackup(&rdbSDK.GetDatabaseBackupRequest{
					DatabaseBackupID: id,
					Region:           region,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("backup (%s) still exists", rs.Primary.ID))
				case httperrors.Is404(err):
					continue
				default:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}

func isBackupPresent(tt *acctest.TestTools, databaseBackup string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[databaseBackup]
		if !ok {
			return fmt.Errorf("resource not found: %s", databaseBackup)
		}

		rdbAPI, region, id, err := rdb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = rdbAPI.GetDatabaseBackup(&rdbSDK.GetDatabaseBackupRequest{
			Region:           region,
			DatabaseBackupID: id,
		})
		if err != nil {
			return fmt.Errorf("failed to get database backup: %w", err)
		}

		return nil
	}
}
