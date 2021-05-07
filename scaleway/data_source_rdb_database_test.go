package scaleway

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mock "github.com/scaleway/terraform-provider-scaleway/scaleway/mocks"
	"github.com/stretchr/testify/assert"
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
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbDatabaseExists(tt, "scaleway_rdb_instance.server", "scaleway_rdb_database.database"),

					resource.TestCheckResourceAttr("data.scaleway_rdb_database.find_by_name_and_instance", "name", "test-terraform"),
					resource.TestCheckResourceAttr("data.scaleway_rdb_database.find_by_name_and_instance", "managed", "true"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database.find_by_name_and_instance", "owner"),
					resource.TestCheckResourceAttrSet("data.scaleway_rdb_database.find_by_name_and_instance", "size"),
				),
			},
		},
	})
}

func TestDataSourceScalewayRdbDatabaseReadWithoutRegionalizedIDUseDefaultResource(t *testing.T) {
	// init testing framework
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	meta, rdbAPI := NewMeta(ctrl)
	data := NewTestResourceDataRawForDataSourceScalewayRDBDatabase(t, "1111-11111111-111111111111")

	// mocking
	rdbAPI.ListDatabasesMustReturnDB("fr-par")

	// run
	diags := dataSourceScalewayRDBDatabaseRead(mock.NewMockContext(ctrl), data, meta)

	// assertions
	assert.Len(diags, 0)
	assertResourceDatabase(assert, data, "fr-par")
}

func TestDataSourceScalewayRdbDatabaseReadWithRegionalizedIDUseDefaultResource(t *testing.T) {
	// init testing framework
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	meta, rdbAPI := NewMeta(ctrl)
	data := NewTestResourceDataRawForDataSourceScalewayRDBDatabase(t, "fr-srr/1111-11111111-111111111111")

	// mocking
	rdbAPI.ListDatabasesMustReturnDB("fr-srr")

	// run
	diags := dataSourceScalewayRDBDatabaseRead(mock.NewMockContext(ctrl), data, meta)

	// assertions
	assert.Len(diags, 0)
	assertResourceDatabase(assert, data, "fr-srr")
}

func NewTestResourceDataRawForDataSourceScalewayRDBDatabase(t *testing.T, uuid string) *schema.ResourceData {
	raw := make(map[string]interface{})
	raw["instance_id"] = uuid
	raw["name"] = "dbname"
	return schema.TestResourceDataRaw(t, dataSourceScalewayRDBDatabase().Schema, raw)
}
