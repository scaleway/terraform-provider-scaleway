package mongodb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	mongodbSDK "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mongodb"
)

func TestAccMongoDBInstance_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-basic1"
						version = "7.0.11"
						node_type = "MGDB-PRO2-XXS"
						node_number = 1
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "name", "test-mongodb-basic1"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "node_type", "mgdb-pro2-xxs"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "version", "7.0.11"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "node_number", "1"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "password", "thiZ_is_v&ry_s3cret"),
				),
			},
		},
	})
}

func TestAccMongoDBInstance_VolumeUpdate(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-volume-update"
						version = "4.4"
						node_type = "db-dev-s"
						node_number = "3"
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						volume_size_in_gb = 50
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "volume_size_in_gb", "50"),
				),
			},
			{
				Config: `
					resource scaleway_mongodb_instance main {
						name = "test-mongodb-volume-update"
						version = "4.4"
						node_type = "db-dev-s"
						node_number = "3"
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						volume_size_in_gb = 100
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "volume_size_in_gb", "100"),
				),
			},
		},
	})
}

func TestAccMongoDBInstance_PrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      IsInstanceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "test-mongodb-private-network"
					}

					resource scaleway_mongodb_instance main {
						name = "test-mongodb-private-network"
						version = "4.4"
						node_type = "db-dev-s"
						node_number = "3"
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						private_network {
							id = scaleway_vpc_private_network.pn01.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isMongoDBInstancePresent(tt, "scaleway_mongodb_instance.main"),
					resource.TestCheckResourceAttr("scaleway_mongodb_instance.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_mongodb_instance.main", "private_network.0.id", "scaleway_vpc_private_network.pn01", "id"),
				),
			},
		},
	})
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
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mongodb_instance" {
				continue
			}

			mongodbAPI, zone, ID, err := mongodb.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			extractRegion, err := zone.Region()
			instance, err := mongodbAPI.GetInstance(&mongodbSDK.GetInstanceRequest{
				InstanceID: ID,
				Region:     extractRegion,
			})
			_ = instance

			if err == nil {
				return fmt.Errorf("instance (%s) still exists", rs.Primary.ID)
			}
			if !httperrors.Is404(err) {
				return err
			}
		}
		return nil
	}
}
