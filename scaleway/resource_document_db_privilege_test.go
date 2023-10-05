package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
)

func TestAccScalewayDocumentDBPrivilege_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayDocumentDBInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_document_db_instance" "instance" {
				  name              = "test-document_db-instance-privilege"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  tags              = ["terraform-test", "scaleway_document_db_instance", "minimal"]
				  volume_size_in_gb = 20
				}

				resource "scaleway_document_db_database" "db01" {
				  instance_id = scaleway_document_db_instance.instance.id
				  name        = "test-document_db-database-basic"
				}

				resource "scaleway_document_db_user" "foo1" {
				  instance_id = scaleway_document_db_instance.instance.id
				  name        = "user_01"
				  password    = "R34lP4sSw#Rd"
				  is_admin    = true
				}

				// Privilege creation with all permission
				resource "scaleway_document_db_privilege" "priv_admin" {
				  instance_id   = scaleway_document_db_instance.instance.id
				  user_name     = scaleway_document_db_user.foo1.name
				  database_name = scaleway_document_db_database.db01.name
				  permission    = "all"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentDBPrivilegeExists(tt, "scaleway_document_db_instance.instance", "scaleway_document_db_database.db01", "scaleway_document_db_user.foo1"),
					resource.TestCheckResourceAttr("scaleway_document_db_privilege.priv_admin", "permission", "all"),
				),
			},
			{
				Config: `
				resource "scaleway_document_db_instance" "instance" {
				  name              = "test-document_db-instance-privilege"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  tags              = ["terraform-test", "scaleway_document_db_instance", "minimal"]
				  volume_size_in_gb = 20
				}

				resource "scaleway_document_db_database" "db01" {
				  instance_id = scaleway_document_db_instance.instance.id
				  name        = "test-document_db-database-basic"
				}

				resource "scaleway_document_db_user" "foo1" {
				  instance_id = scaleway_document_db_instance.instance.id
				  name        = "user_01"
				  password    = "R34lP4sSw#Rd"
				  is_admin    = true
				}

				resource "scaleway_document_db_privilege" "priv_admin" {
				  instance_id   = scaleway_document_db_instance.instance.id
				  user_name     = scaleway_document_db_user.foo1.name
				  database_name = scaleway_document_db_database.db01.name
				  permission    = "all"
				}

				resource "scaleway_document_db_user" "foo2" {
				  instance_id = scaleway_document_db_instance.instance.id
				  name        = "user_02"
				  password    = "R34lP4sSw#Rd"
				}

				// Add new privilege for user foo2 with readwrite permission
				resource "scaleway_document_db_privilege" "priv_foo_02" {
				  instance_id   = scaleway_document_db_instance.instance.id
				  user_name     = scaleway_document_db_user.foo2.name
				  database_name = scaleway_document_db_database.db01.name
				  permission    = "readwrite"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentDBPrivilegeExists(tt, "scaleway_document_db_instance.instance", "scaleway_document_db_database.db01", "scaleway_document_db_user.foo2"),
					resource.TestCheckResourceAttr("scaleway_document_db_privilege.priv_foo_02", "permission", "readwrite"),
				),
			},
			{
				Config: `
			resource "scaleway_document_db_instance" "instance" {
			  name              = "test-document_db-instance-privilege"
			  node_type         = "docdb-play2-pico"
			  engine            = "FerretDB-1"
			  user_name         = "my_initial_user"
			  password          = "thiZ_is_v&ry_s3cret"
			  tags              = ["terraform-test", "scaleway_document_db_instance", "minimal"]
			  volume_size_in_gb = 20
			}

			resource "scaleway_document_db_database" "db01" {
			  instance_id = scaleway_document_db_instance.instance.id
			  name        = "test-document_db-database-basic"
			}

			resource "scaleway_document_db_user" "foo1" {
			  instance_id = scaleway_document_db_instance.instance.id
			  name        = "user_01"
			  password    = "R34lP4sSw#Rd"
			  is_admin    = true
			}

			resource "scaleway_document_db_privilege" "priv_admin" {
			  instance_id   = scaleway_document_db_instance.instance.id
			  user_name     = scaleway_document_db_user.foo1.name
			  database_name = scaleway_document_db_database.db01.name
			  permission    = "all"
			}

			resource "scaleway_document_db_user" "foo2" {
			  instance_id = scaleway_document_db_instance.instance.id
			  name        = "user_02"
			  password    = "R34lP4sSw#Rd"
			}

			resource "scaleway_document_db_privilege" "priv_foo_02" {
			  instance_id   = scaleway_document_db_instance.instance.id
			  user_name     = scaleway_document_db_user.foo2.name
			  database_name = scaleway_document_db_database.db01.name
			  permission    = "readwrite"
			}

			resource "scaleway_document_db_user" "foo3" {
			  instance_id = scaleway_document_db_instance.instance.id
			  name        = "user_03"
			  password    = "R34lP4sSw#Rd"
			}

			// Add a new user privilege with none permission
			resource "scaleway_document_db_privilege" "priv_foo_03" {
			  instance_id   = scaleway_document_db_instance.instance.id
			  user_name     = scaleway_document_db_user.foo3.name
			  database_name = scaleway_document_db_database.db01.name
			  permission    = "none"
			}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentDBPrivilegeExists(tt, "scaleway_document_db_instance.instance", "scaleway_document_db_database.db01", "scaleway_document_db_user.foo3"),
					resource.TestCheckResourceAttr("scaleway_document_db_privilege.priv_foo_03", "permission", "none"),
				),
			},
		},
	})
}

func testAccCheckDocumentDBPrivilegeExists(tt *TestTools, instance string, database string, user string) resource.TestCheckFunc {
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

		api, _, _, err := documentDBAPIWithRegionAndID(tt.Meta, instanceResource.Primary.ID)
		if err != nil {
			return err
		}

		region, instanceID, userName, err := resourceScalewayDocumentDBUserParseID(userResource.Primary.ID)
		if err != nil {
			return err
		}

		_, databaseName, err := resourceScalewayDocumentDBDatabaseName(databaseResource.Primary.ID)
		if err != nil {
			return err
		}

		databases, err := api.ListPrivileges(&documentdb.ListPrivilegesRequest{
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
