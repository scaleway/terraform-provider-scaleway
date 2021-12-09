package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
)

func TestAccScalewayRdbDatabase_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	instanceName := "TestAccScalewayRdbDatabase_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance main {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource scaleway_rdb_database main {
						instance_id = scaleway_rdb_instance.main.id
						name = "foo"
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbDatabaseExists(tt, "scaleway_rdb_instance.main", "scaleway_rdb_database.main"),
					resource.TestCheckResourceAttr("scaleway_rdb_database.main", "name", "foo"),
				),
			},
		},
	})
}

func testAccCheckRdbDatabaseExists(tt *TestTools, instance string, database string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instanceResource, ok := state.RootModule().Resources[instance]
		if !ok {
			return fmt.Errorf("resource not found: %s", instance)
		}

		databaseResource, ok := state.RootModule().Resources[database]
		if !ok {
			return fmt.Errorf("resource database not found: %s", database)
		}

		rdbAPI, _, _, err := rdbAPIWithRegionAndID(tt.Meta, instanceResource.Primary.ID)
		if err != nil {
			return err
		}

		region, instanceID, databaseName, err := resourceScalewayRdbDatabaseParseID(databaseResource.Primary.ID)
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
			return fmt.Errorf("no database found")
		}

		return nil
	}
}

func TestResourceScalewayRdbDatabaseParseIDWithWronglyFormatedIdReturnError(t *testing.T) {
	assert := assert.New(t)
	region, _, _, err := resourceScalewayRdbDatabaseParseID("notandid")
	assert.Error(err)
	assert.Empty(region)
	assert.Equal("can't parse user resource id: notandid", err.Error())
}

func TestResourceScalewayRdbDatabaseParseID(t *testing.T) {
	assert := assert.New(t)
	region, instanceID, dbname, err := resourceScalewayRdbDatabaseParseID("region/instanceid/dbname")
	assert.NoError(err)
	assert.Equal(scw.Region("region"), region)
	assert.Equal("instanceid", instanceID)
	assert.Equal("dbname", dbname)
}
