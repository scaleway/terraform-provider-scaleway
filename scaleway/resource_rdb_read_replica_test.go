package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
)

func TestAccScalewayRdbReadReplica_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-basic"
						node_type = "db-dev-s"
						engine = "PostgreSQL-14"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "minimal" ]
					}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						direct_access {}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.endpoint_id"),
				),
			},
		},
	})
}

func TestAccScalewayRdbReadReplica_PrivateNetwork(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-pn"
						node_type = "db-dev-s"
						engine = "PostgreSQL-14"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "minimal" ]
					}

					resource "scaleway_vpc_private_network" "pn" {}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
							service_ip         = "10.12.1.0/20"
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.endpoint_id"),
				),
			},
		},
	})
}

func TestAccScalewayRdbReadReplica_Update(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-update"
						node_type = "db-dev-s"
						engine = "PostgreSQL-14"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "minimal" ]
					}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						direct_access {}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.endpoint_id"),
				),
			},
			{
				Config: `
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-update"
						node_type = "db-dev-s"
						engine = "PostgreSQL-14"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "minimal" ]
					}

					resource "scaleway_vpc_private_network" "pn" {}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
							service_ip         = "10.12.1.0/20"
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.endpoint_id"),
					resource.TestCheckNoResourceAttr("scaleway_rdb_read_replica.replica", "direct_access.0"),
				),
			},
		},
	})
}

func TestAccScalewayRdbReadReplica_MultipleEndpoints(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-multiple-endpoints"
						node_type = "db-dev-s"
						engine = "PostgreSQL-14"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "minimal" ]
					}

					resource "scaleway_vpc_private_network" "pn" {}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
							service_ip         = "10.12.1.0/20"
						}
						direct_access {}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.endpoint_id"),
				),
			},
		},
	})
}

func testAccCheckRdbReadReplicaExists(tt *TestTools, readReplica string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		readReplicaResource, ok := state.RootModule().Resources[readReplica]
		if !ok {
			return fmt.Errorf("resource not found: %s", readReplica)
		}

		rdbAPI, region, ID, err := rdbAPIWithRegionAndID(tt.Meta, readReplicaResource.Primary.ID)
		if err != nil {
			return err
		}

		_, err = rdbAPI.GetReadReplica(&rdb.GetReadReplicaRequest{
			Region:        region,
			ReadReplicaID: ID,
		})
		if err != nil {
			return fmt.Errorf("can't get read replica: %w", err)
		}

		return nil
	}
}

func testAccCheckScalewayRdbReadReplicaDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_rdb_read_replica" {
				continue
			}

			rdbAPI, region, ID, err := rdbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = rdbAPI.GetReadReplica(&rdb.GetReadReplicaRequest{
				ReadReplicaID: ID,
				Region:        region,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("read_replica (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return fmt.Errorf("error which is not an expected 404: %w", err)
			}
		}

		return nil
	}
}
