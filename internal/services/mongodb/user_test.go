package mongodb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mongodbSDK "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mongodb"
)

func TestAccMongoDBUser_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckMongoDBUserDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name              = "test-mongodb-user"
						version           = "7.0.12"
						node_type         = "MGDB-PLAY2-NANO"
						node_number       = 1
						user_name         = "initial_user"
						password          = "initial_password123"
						volume_size_in_gb = 5
					}

					resource "scaleway_mongodb_user" "main" {
						instance_id = scaleway_mongodb_instance.main.id
						name        = "test_user"
						password    = "test_password123"
						
						roles {
							role          = "read_write"
							database_name = "test_db"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBUserExists(tt, "scaleway_mongodb_user.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_user.main", "name", "test_user"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_user.main", "instance_id"),
					resource.TestCheckResourceAttr("scaleway_mongodb_user.main", "password", "test_password123"),
					resource.TestCheckResourceAttr("scaleway_mongodb_user.main", "roles.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_mongodb_user.main", "roles.*", map[string]string{
						"role":          "read_write",
						"database_name": "test_db",
					}),
				),
			},
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name              = "test-mongodb-user"
						version           = "7.0.12"
						node_type         = "MGDB-PLAY2-NANO"
						node_number       = 1
						user_name         = "initial_user"
						password          = "initial_password123"
						volume_size_in_gb = 5
					}

					resource "scaleway_mongodb_user" "main" {
						instance_id = scaleway_mongodb_instance.main.id
						name        = "test_user"
						password    = "new_password456"
						
						roles {
							role          = "read"
							database_name = "test_db"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBUserExists(tt, "scaleway_mongodb_user.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_user.main", "name", "test_user"),
					resource.TestCheckResourceAttr("scaleway_mongodb_user.main", "password", "new_password456"),
					resource.TestCheckResourceAttr("scaleway_mongodb_user.main", "roles.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_mongodb_user.main", "roles.*", map[string]string{
						"role":          "read",
						"database_name": "test_db",
					}),
				),
			},
		},
	})
}

func TestAccMongoDBUser_StateImport(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckMongoDBUserDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name              = "test-mongodb-user-import"
						version           = "7.0.12"
						node_type         = "MGDB-PLAY2-NANO"
						node_number       = 1
						user_name         = "initial_user"
						password          = "initial_password123"
						volume_size_in_gb = 5
					}

					resource "scaleway_mongodb_user" "main" {
						instance_id = scaleway_mongodb_instance.main.id
						name        = "import_user"
						password    = "import_password123"
						
						roles {
							role          = "db_admin"
							database_name = "admin"
						}
						
						roles {
							role         = "read"
							any_database = true
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBUserExists(tt, "scaleway_mongodb_user.main"),
				),
			},
			{
				ResourceName:      "scaleway_mongodb_user.main",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore roles: API reorders/normalizes and TypeSet flattens by index: flaky import diff.
				// TODO: add deterministic sorting or a set-aware StateCheck to verify roles.
				ImportStateVerifyIgnore: []string{"password", "roles"},
			},
		},
	})
}

func testAccCheckMongoDBUserExists(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		region, instanceID, userName, err := mongodb.ResourceUserParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		mongodbAPI := mongodbSDK.NewAPI(tt.Meta.ScwClient())

		res, err := mongodbAPI.ListUsers(&mongodbSDK.ListUsersRequest{
			Region:     region,
			InstanceID: instanceID,
			Name:       &userName,
		})
		if err != nil {
			return err
		}

		if len(res.Users) == 0 {
			return fmt.Errorf("MongoDB user %s not found", userName)
		}

		return nil
	}
}

func testAccCheckMongoDBUserDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := mongodbSDK.NewAPI(tt.Meta.ScwClient())
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range s.RootModule().Resources {
				if rs.Type != "scaleway_mongodb_user" {
					continue
				}

				region, instanceID, userName, err := mongodb.ResourceUserParseID(rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				res, err := api.ListUsers(&mongodbSDK.ListUsersRequest{
					Region:     region,
					InstanceID: instanceID,
					Name:       &userName,
				})

				switch {
				case err == nil && len(res.Users) > 0:
					return retry.RetryableError(fmt.Errorf("MongoDB user %s still exists", userName))
				case err == nil && len(res.Users) == 0:
					continue
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
