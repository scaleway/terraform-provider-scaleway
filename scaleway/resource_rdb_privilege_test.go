package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
)

func TestAccScalewayRdbPrivilege_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	instanceName := "TestAccScalewayRdbPrivilege_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "instance" {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource "scaleway_rdb_database" "db01" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
					}

					resource "scaleway_rdb_user" "foo1" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "user_01"
						password = "R34lP4sSw#Rd"
						is_admin = true
					}

					resource "scaleway_rdb_privilege" "priv_admin" {
						instance_id   = scaleway_rdb_instance.instance.id
						user_name     = scaleway_rdb_user.foo1.name
						database_name = scaleway_rdb_database.db01.name
						permission    = "all"
					}
					`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbPrivilegeExists(tt, "scaleway_rdb_instance.instance", "scaleway_rdb_database.db01", "scaleway_rdb_user.foo1"),
					resource.TestCheckResourceAttr("scaleway_rdb_privilege.priv_admin", "permission", "all"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "instance" {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource "scaleway_rdb_database" "db01" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
					}

					resource "scaleway_rdb_user" "foo1" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "user_01"
						password = "R34lP4sSw#Rd"
						is_admin = true
					}

					resource "scaleway_rdb_privilege" "priv_admin" {
						instance_id   = scaleway_rdb_instance.instance.id
						user_name     = scaleway_rdb_user.foo1.name
						database_name = scaleway_rdb_database.db01.name
						permission    = "all"
					}

					resource "scaleway_rdb_user" "foo2" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "user_02"
						password = "R34lP4sSw#Rd"
					}

					resource "scaleway_rdb_privilege" "priv_foo_02" {
						instance_id   = scaleway_rdb_instance.instance.id
						user_name     = scaleway_rdb_user.foo2.name
						database_name = scaleway_rdb_database.db01.name
						permission    = "readwrite"
					}
					`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbPrivilegeExists(tt, "scaleway_rdb_instance.instance", "scaleway_rdb_database.db01", "scaleway_rdb_user.foo2"),
					resource.TestCheckResourceAttr("scaleway_rdb_privilege.priv_foo_02", "permission", "readwrite"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_rdb_instance" "instance" {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource "scaleway_rdb_database" "db01" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "foo"
					}

					resource "scaleway_rdb_user" "foo1" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "user_01"
						password = "R34lP4sSw#Rd"
						is_admin = true
					}

					resource "scaleway_rdb_privilege" "priv_admin" {
						instance_id   = scaleway_rdb_instance.instance.id
						user_name     = scaleway_rdb_user.foo1.name
						database_name = scaleway_rdb_database.db01.name
						permission    = "all"
					}

					resource "scaleway_rdb_user" "foo2" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "user_02"
						password = "R34lP4sSw#Rd"
					}

					resource "scaleway_rdb_privilege" "priv_foo_02" {
						instance_id   = scaleway_rdb_instance.instance.id
						user_name     = scaleway_rdb_user.foo2.name
						database_name = scaleway_rdb_database.db01.name
						permission    = "readwrite"
					}

					resource "scaleway_rdb_user" "foo3" {
						instance_id = scaleway_rdb_instance.instance.id
						name = "user_03"
						password = "R34lP4sSw#Rd"
					}

					resource "scaleway_rdb_privilege" "priv_foo_03" {
						instance_id   = scaleway_rdb_instance.instance.id
						user_name     = scaleway_rdb_user.foo3.name
						database_name = scaleway_rdb_database.db01.name
						permission    = "none"
					}
					`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbPrivilegeExists(tt, "scaleway_rdb_instance.instance", "scaleway_rdb_database.db01", "scaleway_rdb_user.foo3"),
					resource.TestCheckResourceAttr("scaleway_rdb_privilege.priv_foo_03", "permission", "none"),
				),
			},
		},
	})
}

func TestAccScalewayRdbPrivilege_ConcurrentPrivileges(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	instanceName := "TestAccScalewayRdbPrivilege_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					locals {
						permissions = {
							"application-default" = {
								user_name     = "application"
								database_name = "default"
								permission    = "all"
							},
							"application-core" = {
								user_name     = "application"
								database_name = "core"
								permission    = "all"
							},
							"user-core" = {
								user_name     = "user"
								database_name = "core"
								permission    = "readonly"
							}
						}
					}

					resource "scaleway_rdb_instance" "main" {
						name           = "%s"
						node_type      = "DB-DEV-S"
						engine         = "PostgreSQL-11"
						disable_backup = true
						user_name      = "my_initial_user"
						password       = "thiZ_is_v&ry_s3cret"
					}

					resource "scaleway_rdb_database" "databases" {
						instance_id = scaleway_rdb_instance.main.id
						name        = "core"
					}

					resource "scaleway_rdb_database" "default" {
						instance_id = scaleway_rdb_instance.main.id
						name        = "default"
					}

					resource "scaleway_rdb_user" "user" {
						instance_id = scaleway_rdb_instance.main.id
						name = "user"
						password = "thiZ_is_v&ry_s3cret"
					}
					
					resource "scaleway_rdb_user" "application" {
						instance_id = scaleway_rdb_instance.main.id
						name = "application"
						password = "thiZ_is_v&ry_s3cret"
					}
					
					resource "scaleway_rdb_privilege" "permission" {
						for_each = local.permissions
					
						instance_id = scaleway_rdb_instance.main.id
					
						user_name     = each.value.user_name
						database_name = each.value.database_name
						permission    = each.value.permission
					
						depends_on = [scaleway_rdb_user.user, scaleway_rdb_database.databases]
					}
					`, instanceName),
				//Check: resource.ComposeTestCheckFunc(
				//	testAccCheckRdbPrivilegeExists(tt, "scaleway_rdb_instance.instance", "scaleway_rdb_database.db01", "scaleway_rdb_user.foo1"),
				//	resource.TestCheckResourceAttr("scaleway_rdb_privilege.priv_admin", "permission", "all"),
				//),
			},
		},
	})
}

func testAccCheckRdbPrivilegeExists(tt *TestTools, instance string, database string, user string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instanceResource, ok := state.RootModule().Resources[instance]
		if !ok {
			return fmt.Errorf("resource not found: %s", instance)
		}

		databaseResource, ok := state.RootModule().Resources[database]
		if !ok {
			return fmt.Errorf("resource database not found: %s", database)
		}

		userResource, ok := state.RootModule().Resources[user]
		if !ok {
			return fmt.Errorf("resource not found: %s", user)
		}

		rdbAPI, _, _, err := rdbAPIWithRegionAndID(tt.Meta, instanceResource.Primary.ID)
		if err != nil {
			return err
		}

		_, _, databaseName, err := resourceScalewayRdbDatabaseParseID(databaseResource.Primary.ID)
		if err != nil {
			return err
		}

		region, instanceID, userName, err := resourceScalewayRdbUserParseID(userResource.Primary.ID)
		if err != nil {
			return err
		}

		databases, err := rdbAPI.ListPrivileges(&rdb.ListPrivilegesRequest{
			Region:       region,
			InstanceID:   instanceID,
			DatabaseName: &databaseName,
			UserName:     &userName,
		})
		if err != nil {
			return err
		}

		if len(databases.Privileges) != 1 {
			return fmt.Errorf("no privilege found")
		}

		return nil
	}
}
