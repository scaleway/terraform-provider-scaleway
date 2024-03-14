package scaleway_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccScalewayRdbDatabase_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := testAccCheckScalewayRdbEngineGetLatestVersion(tt, postgreSQLEngineName)

	instanceName := "TestAccScalewayRdbDatabase_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance main {
						name = "%s"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource scaleway_rdb_database main {
						instance_id = scaleway_rdb_instance.main.id
						name = "foo"
					}`, instanceName, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbDatabaseExists(tt, "scaleway_rdb_instance.main", "scaleway_rdb_database.main"),
					resource.TestCheckResourceAttr("scaleway_rdb_database.main", "name", "foo"),
				),
			},
		},
	})
}

func TestAccScalewayRdbDatabase_ManualDelete(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := testAccCheckScalewayRdbEngineGetLatestVersion(tt, postgreSQLEngineName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "pgsql" {
						name           = "bug"
						node_type      = "db-dev-m"
						engine         = %q
						is_ha_cluster  = false
						disable_backup = true
						user_name      = "admin"
						password       = "thiZ_is_v&ry_s3cret"
						tags           = ["bug"]
					}

					resource "scaleway_rdb_user" "bug" {
						instance_id = scaleway_rdb_instance.pgsql.id
						name        = "bug"
						password    = "thiZ_is_v&ry_s3cret"
						is_admin    = false
					}

					resource "scaleway_rdb_database" "bug" {
						instance_id = scaleway_rdb_instance.pgsql.id
						name        = "bug"
					}

					resource "scaleway_rdb_privilege" "bug" {
						instance_id   = scaleway_rdb_instance.pgsql.id
						user_name     = "bug"
						database_name = "bug"
						permission    = "all"

						depends_on = [scaleway_rdb_user.bug, scaleway_rdb_database.bug]
					}
				`, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbDatabaseExists(tt, "scaleway_rdb_instance.pgsql", "scaleway_rdb_database.bug"),
					resource.TestCheckResourceAttr("scaleway_rdb_database.bug", "name", "bug"),
				),
			},
		},
	})
}

func testAccCheckRdbDatabaseExists(tt *acctest.TestTools, instance string, database string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instanceResource, ok := state.RootModule().Resources[instance]
		if !ok {
			return fmt.Errorf("resource not found: %s", instance)
		}

		databaseResource, ok := state.RootModule().Resources[database]
		if !ok {
			return fmt.Errorf("resource database not found: %s", database)
		}

		rdbAPI, _, _, err := scaleway.RdbAPIWithRegionAndID(tt.Meta, instanceResource.Primary.ID)
		if err != nil {
			return err
		}

		region, instanceID, databaseName, err := scaleway.ResourceScalewayRdbDatabaseParseID(databaseResource.Primary.ID)
		if err != nil {
			return err
		}

		databases, err := rdbAPI.ListDatabases(&rdb.ListDatabasesRequest{
			Region:     region,
			InstanceID: instanceID,
			Name:       &databaseName,
			Managed:    nil,
			Owner:      nil,
			OrderBy:    "",
		})
		if err != nil {
			return err
		}

		if len(databases.Databases) != 1 {
			return errors.New("no database found")
		}

		return nil
	}
}

func TestResourceScalewayRdbDatabaseParseIDWithWronglyFormatedIdReturnError(t *testing.T) {
	region, _, _, err := scaleway.ResourceScalewayRdbDatabaseParseID("notandid")
	require.Error(t, err)
	assert.Empty(t, region)
	assert.Equal(t, "can't parse user resource id: notandid", err.Error())
}

func TestResourceScalewayRdbDatabaseParseID(t *testing.T) {
	region, instanceID, dbname, err := scaleway.ResourceScalewayRdbDatabaseParseID("region/instanceid/dbname")
	require.NoError(t, err)
	assert.Equal(t, scw.Region("region"), region)
	assert.Equal(t, "instanceid", instanceID)
	assert.Equal(t, "dbname", dbname)
}
