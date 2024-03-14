package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccScalewayDataSourceRdbPrivilege_Basic(t *testing.T) {
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
					resource "scaleway_rdb_instance" "instance" {
						name = "test-privilege"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}`, latestEngineVersion),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "instance" {
						name = "test-privilege"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource "scaleway_rdb_database" "db" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
					}`, latestEngineVersion),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "instance" {
						name = "test-privilege"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource "scaleway_rdb_database" "db" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
					}

					resource "scaleway_rdb_user" "foo" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
						password = "R34lP4sSw#Rd"
					}`, latestEngineVersion),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "instance" {
						name = "test-privilege"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource "scaleway_rdb_database" "db" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
					}

					resource "scaleway_rdb_user" "foo" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
						password = "R34lP4sSw#Rd"
					}

					resource "scaleway_rdb_privilege" "priv" {
						instance_id   = scaleway_rdb_instance.instance.id
						user_name     = scaleway_rdb_user.foo.name
						database_name = scaleway_rdb_database.db.name
						permission    = "all"
					}`, latestEngineVersion),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "instance" {
						name = "test-privilege"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource "scaleway_rdb_database" "db" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
					}

					resource "scaleway_rdb_user" "foo" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
						password = "R34lP4sSw#Rd"
					}

					resource "scaleway_rdb_privilege" "priv" {
						instance_id   = scaleway_rdb_instance.instance.id
						user_name     = scaleway_rdb_user.foo.name
						database_name = scaleway_rdb_database.db.name
						permission    = "all"
					}

					data "scaleway_rdb_privilege" "find_priv" {
						instance_id   = scaleway_rdb_instance.instance.id
						user_name     = scaleway_rdb_user.foo.name
						database_name = scaleway_rdb_database.db.name
					}
				`, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbDatabaseExists(tt, "scaleway_rdb_instance.instance", "scaleway_rdb_database.db"),
					resource.TestCheckResourceAttr("data.scaleway_rdb_privilege.find_priv", "permission", "all"),
					resource.TestCheckResourceAttr("data.scaleway_rdb_privilege.find_priv", "region", "fr-par"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "instance" {
						name = "test-privilege"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}`, latestEngineVersion),
			},
		},
	})
}
