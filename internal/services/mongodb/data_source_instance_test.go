package mongodb_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mongodbSDK "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func TestAccDataSourceMongoDBInstance_Lookups(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mongodb_instance" "test" {
						name        = "test-mongodb-instance-by-name"
						version     = "7.0"
						node_type   = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name   = "my_initial_user"
						password    = "thiZ_is_v&ry_s3cret"
					}

					data "scaleway_mongodb_instance" "test_by_name" {
						name = scaleway_mongodb_instance.test.name
					}

					data "scaleway_mongodb_instance" "test_by_id" {
						instance_id = scaleway_mongodb_instance.test.id
					}

					data "scaleway_mongodb_databases" "main" {
						instance_id = scaleway_mongodb_instance.test.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.test_by_name", "name", "test-mongodb-instance-by-name"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.test_by_name", "version", "7.0"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.test_by_name", "node_type", "mgdb-play2-nano"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.test_by_name", "node_number", "1"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.test_by_id", "name", "test-mongodb-instance-by-name"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.test_by_id", "version", "7.0"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.test_by_id", "node_type", "mgdb-play2-nano"),
					resource.TestCheckResourceAttr("data.scaleway_mongodb_instance.test_by_id", "node_number", "1"),
					checkMongoDBDatabasesMatchAPI(tt, "data.scaleway_mongodb_databases.main"),
				),
			},
		},
	})
}

func checkMongoDBDatabasesMatchAPI(tt *acctest.TestTools, dataSourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("not found: %s", dataSourceName)
		}

		region, instanceID, err := regional.ParseID(rs.Primary.Attributes["instance_id"])
		if err != nil {
			return fmt.Errorf("failed to parse instance ID: %w", err)
		}

		api := mongodbSDK.NewAPI(tt.Meta.ScwClient())

		res, err := api.ListDatabases(&mongodbSDK.ListDatabasesRequest{
			Region:     region,
			InstanceID: instanceID,
			OrderBy:    mongodbSDK.ListDatabasesRequestOrderByNameAsc,
		}, scw.WithContext(tt.T.Context()), scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list databases: %w", err)
		}

		stateCount := 0
		if stateCountStr, ok := rs.Primary.Attributes["databases.#"]; ok && stateCountStr != "" {
			stateCount, err = strconv.Atoi(stateCountStr)
			if err != nil {
				return fmt.Errorf("failed to parse databases count: %w", err)
			}
		}

		if stateCount != len(res.Databases) {
			return fmt.Errorf("expected %d databases in state, got %d", len(res.Databases), stateCount)
		}

		for i, database := range res.Databases {
			stateName := rs.Primary.Attributes[fmt.Sprintf("databases.%d.name", i)]
			if stateName != database.Name {
				return fmt.Errorf("expected database %q at index %d, got %q", database.Name, i, stateName)
			}
		}

		return nil
	}
}
