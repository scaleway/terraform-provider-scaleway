package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/errs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func init() {
	resource.AddTestSweepers("scaleway_rdb_database_backup", &resource.Sweeper{
		Name: "scaleway_rdb_database_backup",
		F:    testSweepRDBDatabaseBackup,
	})
}

func testSweepRDBDatabaseBackup(_ string) error {
	return sweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		rdbAPI := rdb.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the rdb database backups in (%s)", region)
		listBackups, err := rdbAPI.ListDatabaseBackups(&rdb.ListDatabaseBackupsRequest{
			Region: region,
		})
		if err != nil {
			return fmt.Errorf("error listing rdb database backups in (%s) in sweeper: %s", region, err)
		}

		for _, backup := range listBackups.DatabaseBackups {
			_, err := rdbAPI.DeleteDatabaseBackup(&rdb.DeleteDatabaseBackupRequest{
				Region:           region,
				DatabaseBackupID: backup.ID,
			})
			if err != nil && !errs.Is404Error(err) {
				return fmt.Errorf("error deleting rdb database backup in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayRdbDatabaseBackup_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	instanceName := "TestAccScalewayRdbDatabaseBackup_Basic"
	latestEngineVersion := testAccCheckScalewayRdbEngineGetLatestVersion(tt, postgreSQLEngineName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbDatabaseBackupDestroy(tt),
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
					testAccCheckRdbDatabaseExists(tt, "scaleway_rdb_instance.main", "scaleway_rdb_database.main"),
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
					testAccCheckRdbDatabaseBackupExists(tt, "scaleway_rdb_database_backup.main"),

					resource.TestCheckResourceAttr("scaleway_rdb_database_backup.main", "database_name", "foo"),
					resource.TestCheckResourceAttr("scaleway_rdb_database_backup.main", "name", "test_backup"),
				),
			},
		},
	})
}

func testAccCheckScalewayRdbDatabaseBackupDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_rdb_database_backup" {
				continue
			}

			rdbAPI, region, ID, err := rdbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = rdbAPI.GetDatabaseBackup(&rdb.GetDatabaseBackupRequest{
				DatabaseBackupID: ID,
				Region:           region,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("backup (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !errs.Is404Error(err) {
				return err
			}
		}

		return nil
	}
}

func testAccCheckRdbDatabaseBackupExists(tt *TestTools, databaseBackup string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[databaseBackup]
		if !ok {
			return fmt.Errorf("resource not found: %s", databaseBackup)
		}

		rdbAPI, region, id, err := rdbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = rdbAPI.GetDatabaseBackup(&rdb.GetDatabaseBackupRequest{
			Region:           region,
			DatabaseBackupID: id,
		})
		if err != nil {
			return fmt.Errorf("failed to get database backup: %w", err)
		}

		return nil
	}
}
