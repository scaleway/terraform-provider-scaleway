package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mongodbSDK "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mongodb"
)

var DestroyWaitTimeout = 3 * time.Minute

func TestAccMongoDBInstance_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-basic-1"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "name", "test-mongodb-basic-1"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "node_type", "mgdb-play2-nano"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "version", "7.0"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "node_number", "1"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "tls_certificate"),
				),
			},
		},
	})
}

func TestAccMongoDBInstance_VolumeUpdate(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-volume-update1"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = "1"
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						volume_size_in_gb = 5
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "volume_size_in_gb", "5"),
				),
			},
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-volume-update1"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = "1"
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						volume_size_in_gb = 10
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "volume_size_in_gb", "10"),
				),
			},
		},
	})
}

func TestAccMongoDBInstance_SnapshotSchedule(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-snapshot-schedule"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						snapshot_schedule_frequency_hours = 24
						snapshot_schedule_retention_days = 7
						is_snapshot_schedule_enabled = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "snapshot_schedule_frequency_hours", "24"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "snapshot_schedule_retention_days", "7"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "is_snapshot_schedule_enabled", "true"),
				),
			},
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-snapshot-schedule"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						snapshot_schedule_frequency_hours = 12
						snapshot_schedule_retention_days = 14
						is_snapshot_schedule_enabled = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "snapshot_schedule_frequency_hours", "12"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "snapshot_schedule_retention_days", "14"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "is_snapshot_schedule_enabled", "false"),
				),
			},
		},
	})
}

func TestAccMongoDBInstance_UpdateNameTagsUser(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-update-initial"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name = "user"
						password = "initial_password"
						tags = ["initial_tag1", "initial_tag2"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "name", "test-mongodb-update-initial"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "user_name", "user"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "tags.0", "initial_tag1"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "tags.1", "initial_tag2"),
				),
			},
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-update-final"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name = "user"
						password = "updated_password"
						tags = ["updated_tag1", "updated_tag2", "updated_tag3"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "name", "test-mongodb-update-final"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "tags.#", "3"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "tags.0", "updated_tag1"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "tags.1", "updated_tag2"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "tags.2", "updated_tag3"),
				),
			},
		},
	})
}

func TestAccMongoDBInstance_FromSnapshot(t *testing.T) {
	t.Skip("TestAccMongoDBInstance_FromSnapshot skipped: waiting for stability fix from database team.")

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name        = "test-mongodb-from-snapshot"
						version     = "7.0.12"
						node_type   = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name   = "my_initial_user"
						password    = "thiZ_is_v&ry_s3cret"
					}

					resource "scaleway_mongodb_snapshot" "main_snapshot" {
						instance_id = scaleway_mongodb_instance.main.id
						name        = "test-snapshot"
						expires_at  = "2026-03-20T23:59:59Z"
 						depends_on = [
    						scaleway_mongodb_instance.main
  						]
					}
					
					resource "scaleway_mongodb_instance" "restored_instance" {
						snapshot_id = scaleway_mongodb_snapshot.main_snapshot.id
						name        = "restored-mongodb-from-snapshot"
						node_type   = "MGDB-PLAY2-NANO"
						node_number = 1
						depends_on = [
    						scaleway_mongodb_instance.main,
    						scaleway_mongodb_snapshot.main_snapshot
  						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.restored_instance"),
				),
			},
		},
	})
}

func TestAccMongoDBInstance_WithPrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc main {
						region = "fr-par"
						name = "TestAccMongoDBInstance_WithPrivateNetwork"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "my_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_vpc_private_network.pn01", "name", "my_private_network"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						region = "fr-par"
						name = "TestAccMongoDBInstance_WithPrivateNetwork"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "my_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_mongodb_instance main {
						name = "test-mongodb-private-network"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						private_network {
							pn_id = "${scaleway_vpc_private_network.pn01.id}"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.dns_records.0"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_ip.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_ip.0.address"),
				),
			},
		},
	})
}

func TestAccMongoDBInstance_UpdatePrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var previousPrivateNetworkID string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc main {
						region = "fr-par"
						name = "TestAccMongoDBInstance_UpdatePrivateNetwork"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "my_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_vpc_private_network pn02 {
						name = "update_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_vpc_private_network.pn01", "name", "my_private_network"),
					resource.TestCheckResourceAttr("scaleway_vpc_private_network.pn02", "name", "update_private_network"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						region = "fr-par"
						name = "TestAccMongoDBInstance_UpdatePrivateNetwork"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "my_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_vpc_private_network pn02 {
						name = "update_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_mongodb_instance main {
						name = "test-mongodb-private-network"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						private_network {
							pn_id = "${scaleway_vpc_private_network.pn01.id}"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					capturePrivateNetworkID(&previousPrivateNetworkID),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_ip.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_ip.0.address"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						region = "fr-par"
						name = "TestAccMongoDBInstance_UpdatePrivateNetwork"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "my_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_vpc_private_network pn02 {
						name = "update_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_mongodb_instance main {
						name = "test-mongodb-private-network"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						private_network {
							pn_id = "${scaleway_vpc_private_network.pn02.id}"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					verifyPrivateNetworkIDChanged(&previousPrivateNetworkID),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_ip.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_ip.0.address"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						region = "fr-par"
						name = "TestAccMongoDBInstance_UpdatePrivateNetwork"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "my_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_vpc_private_network pn02 {
						name = "update_private_network"
						region = "fr-par"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_mongodb_instance main {
						name = "test-mongodb-private-network"
						version = "7.0.12"
						node_type = "MGDB-PLAY2-NANO"
						node_number = 1
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckNoResourceAttr("scaleway_mongodb_instance.main", "private_network.#"),
				),
			},
		},
	})
}

func TestAccMongoDBInstance_WithPublicNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_vpc" "main" {
					name = "TestAccMongoDBInstance_WithPublicNetwork"
					region = "fr-par"
				}

				resource "scaleway_vpc_private_network" "pn01" {
  					name   = "my_private_network"
  					region = "fr-par"
				}

				resource "scaleway_mongodb_instance" "main" {
				  name              = "test-mongodb-public-network"
				  version           = "7.0.12"
				  node_type         = "MGDB-PLAY2-NANO"
				  node_number       = 1
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 5

				  private_network {
				    pn_id = scaleway_vpc_private_network.pn01.id
				  }

				  public_network {}
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "public_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "public_network.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "public_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "public_network.0.dns_record"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_mongodb_instance.main", "private_network.0.dns_records.#"),
				),
			},
		},
	})
}

func capturePrivateNetworkID(previousID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["scaleway_mongodb_instance.main"]
		if !ok {
			return errors.New("MongoDB instance not found in state")
		}

		id, ok := rs.Primary.Attributes["private_network.0.id"]
		if !ok {
			return errors.New("private_network.0.id attribute not found")
		}

		*previousID = id

		return nil
	}
}

func verifyPrivateNetworkIDChanged(previousID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["scaleway_mongodb_instance.main"]
		if !ok {
			return errors.New("MongoDB instance not found in state")
		}

		newID, ok := rs.Primary.Attributes["private_network.0.id"]
		if !ok {
			return errors.New("private_network.0.id attribute not found")
		}

		if *previousID == "" {
			return errors.New("previousPrivateNetworkID was not set in previous step")
		}

		if *previousID == newID {
			return fmt.Errorf("expected private_network.0.id to change, but it remained the same: %s", newID)
		}

		return nil
	}
}

func isMongoDBInstancePresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		mongodbAPI, region, ID, err := mongodb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = mongodbAPI.GetInstance(&mongodbSDK.GetInstanceRequest{
			InstanceID: ID,
			Region:     region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func IsInstanceDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_mongodb_instance" {
					continue
				}

				api, region, id, err := mongodb.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetInstance(&mongodbSDK.GetInstanceRequest{
					InstanceID: id,
					Region:     region,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("mongodb instance (%s) still exists", rs.Primary.ID))
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

func TestAccMongoDBInstance_PasswordWO(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name              = "test-mongodb-password-wo"
						version           = "7.0.12"
						node_type         = "MGDB-PLAY2-NANO"
						node_number       = 1
						user_name         = "my_initial_user"
						password_wo       = "thiZ_is_v&ry_s3cret_WO_1"
						password_wo_version = 1
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "name", "test-mongodb-password-wo"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "node_type", "mgdb-play2-nano"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "version", "7.0"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "node_number", "1"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "user_name", "my_initial_user"),
					// password_wo should not be set in state
					resource.TestCheckNoResourceAttr("scaleway_mongodb_instance.main", "password_wo"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "password_wo_version", "1"),
				),
			},
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name              = "test-mongodb-password-wo"
						version           = "7.0.12"
						node_type         = "MGDB-PLAY2-NANO"
						node_number       = 1
						user_name         = "my_initial_user"
						password_wo       = "thiZ_is_v&ry_s3cret_WO_2"
						password_wo_version = 2
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "password_wo_version", "2"),
				),
			},
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name              = "test-mongodb-password-wo"
						version           = "7.0.12"
						node_type         = "MGDB-PLAY2-NANO"
						node_number       = 1
						user_name         = "my_initial_user"
						password          = "thiZ_is_v&ry_s3cret_regular"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "password", "thiZ_is_v&ry_s3cret_regular"),
				),
			},
			{
				Config: `
					resource "scaleway_mongodb_instance" "main" {
						name              = "test-mongodb-password-wo"
						version           = "7.0.12"
						node_type         = "MGDB-PLAY2-NANO"
						node_number       = 1
						user_name         = "my_initial_user"
						password_wo       = "thiZ_is_v&ry_s3cret_WO_final"
						password_wo_version = 3
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "password_wo_version", "3"),
				),
			},
		},
	})
}
