package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	serverless_sqldb "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func init() {
	resource.AddTestSweepers("scaleway_sdb_sql_database", &resource.Sweeper{
		Name: "scaleway_sdb_sql_database",
		F:    testSweepServerlessSQLDBDatabase,
	})
}

func testSweepServerlessSQLDBDatabase(_ string) error {
	return sweepRegions((&serverless_sqldb.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		sdbAPI := serverless_sqldb.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the serverless sql database in (%s)", region)
		listServerlessSQLDBDatabases, err := sdbAPI.ListDatabases(
			&serverless_sqldb.ListDatabasesRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing database in (%s) in sweeper: %s", region, err)
		}

		for _, database := range listServerlessSQLDBDatabases.Databases {
			_, err := sdbAPI.DeleteDatabase(&serverless_sqldb.DeleteDatabaseRequest{
				DatabaseID: database.ID,
				Region:     region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting database in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayServerlessSQLDBDatabase_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayServerlessSQLDBDatabaseDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_sdb_sql_database main {
						name = "test-sdb-sql-database-basic"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerlessSQLDBDatabaseExists(tt, "scaleway_sdb_sql_database.main"),
					testCheckResourceAttrUUID("scaleway_sdb_sql_database.main", "id"),
					resource.TestCheckResourceAttr("scaleway_sdb_sql_database.main", "name", "test-sdb-sql-database-basic"),
					resource.TestCheckResourceAttr("scaleway_sdb_sql_database.main", "min_cpu", "0"),
					resource.TestCheckResourceAttr("scaleway_sdb_sql_database.main", "max_cpu", "15"),
					resource.TestCheckResourceAttrSet("scaleway_sdb_sql_database.main", "endpoint"),
				),
			},
			{
				Config: `
					resource scaleway_sdb_sql_database main {
						name = "test-sdb-sql-database-basic"
						max_cpu = 6
						min_cpu = 2
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerlessSQLDBDatabaseExists(tt, "scaleway_sdb_sql_database.main"),
					testCheckResourceAttrUUID("scaleway_sdb_sql_database.main", "id"),
					resource.TestCheckResourceAttr("scaleway_sdb_sql_database.main", "min_cpu", "2"),
					resource.TestCheckResourceAttr("scaleway_sdb_sql_database.main", "max_cpu", "6"),
				),
			},
			{
				Config: `
					resource scaleway_sdb_sql_database main {
						name = "test-sdb-sql-database-basic"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerlessSQLDBDatabaseExists(tt, "scaleway_sdb_sql_database.main"),
					testCheckResourceAttrUUID("scaleway_sdb_sql_database.main", "id"),
					resource.TestCheckResourceAttr("scaleway_sdb_sql_database.main", "min_cpu", "0"),
					resource.TestCheckResourceAttr("scaleway_sdb_sql_database.main", "max_cpu", "15"),
				),
			},
			{ // Should ForceNew, this is testing creation with cpu values
				Config: `
					resource scaleway_sdb_sql_database main {
						name = "test-sdb-sql-database-basic-rename"
						min_cpu = 4
						max_cpu = 8
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerlessSQLDBDatabaseExists(tt, "scaleway_sdb_sql_database.main"),
					testCheckResourceAttrUUID("scaleway_sdb_sql_database.main", "id"),
					resource.TestCheckResourceAttr("scaleway_sdb_sql_database.main", "min_cpu", "4"),
					resource.TestCheckResourceAttr("scaleway_sdb_sql_database.main", "max_cpu", "8"),
				),
			},
		},
	})
}

func testAccCheckScalewayServerlessSQLDBDatabaseExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := serverlessSQLdbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDatabase(&serverless_sqldb.GetDatabaseRequest{
			DatabaseID: id,
			Region:     region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayServerlessSQLDBDatabaseDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_sdb_sql_database" {
				continue
			}

			api, region, id, err := serverlessSQLdbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteDatabase(&serverless_sqldb.DeleteDatabaseRequest{
				DatabaseID: id,
				Region:     region,
			})

			if err == nil {
				return fmt.Errorf("serverless_sql database (%s) still exists", rs.Primary.ID)
			}

			if !is403Error(err) {
				return err
			}
		}

		return nil
	}
}
