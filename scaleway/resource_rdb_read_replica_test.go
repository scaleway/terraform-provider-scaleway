package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func TestAccScalewayRdbReadReplica_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := testAccCheckScalewayRdbEngineGetLatestVersion(tt, postgreSQLEngineName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-basic"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "minimal" ]
					}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						direct_access {}
					}`, latestEngineVersion),
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
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := testAccCheckScalewayRdbEngineGetLatestVersion(tt, postgreSQLEngineName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-pn"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "private-network" ]
					}

					resource "scaleway_vpc_private_network" "pn" {}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
							service_ip         = "10.12.1.0/20"
						}
					}`, latestEngineVersion),
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
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := testAccCheckScalewayRdbEngineGetLatestVersion(tt, postgreSQLEngineName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-update"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "update" ]
					}

					resource "scaleway_rdb_read_replica" "replica" {
						instance_id = scaleway_rdb_instance.instance.id
						direct_access {}
					}`, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "direct_access.0.endpoint_id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-update"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "update" ]
					}

					resource "scaleway_vpc_private_network" "pn" {}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
							service_ip         = "10.12.1.0/20"
						}
					}`, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "direct_access.#", "0"),
				),
			},
			// Keep PN but change static config -> ipam
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-update"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "update" ]
					}

					resource "scaleway_vpc_private_network" "pn" {}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
							enable_ipam = true
						}
					}`, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "private_network.0.private_network_id", "scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "direct_access.#", "0"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.#", "1"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.0.enable_ipam", "true"),
				),
			},
			// Change PN but keep ipam config
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-update"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "update" ]
					}
			
					resource "scaleway_vpc_private_network" "pn" {}
					resource "scaleway_vpc_private_network" "pn2" {}
			
					resource "scaleway_rdb_read_replica" "replica" {
						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn2.id
							enable_ipam = true
						}
					}`, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "private_network.0.private_network_id", "scaleway_vpc_private_network.pn2", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "direct_access.#", "0"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.#", "1"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.0.enable_ipam", "true"),
				),
			},
			// Keep PN but change ipam config -> static
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-update"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "update" ]
					}
			
					resource "scaleway_vpc_private_network" "pn" {}
					resource "scaleway_vpc_private_network" "pn2" {}
			
					resource "scaleway_rdb_read_replica" "replica" {
						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn2.id
							service_ip = "10.12.1.0/20"
						}
					}`, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "private_network.0.private_network_id", "scaleway_vpc_private_network.pn2", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "direct_access.#", "0"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.#", "1"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.0.enable_ipam", "false"),
				),
			},
			// Change PN but keep static config
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-update"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "update" ]
					}
			
					resource "scaleway_vpc_private_network" "pn" {}
					resource "scaleway_vpc_private_network" "pn2" {}
			
					resource "scaleway_rdb_read_replica" "replica" {
						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
							service_ip = "10.12.1.0/20"
						}
					}`, latestEngineVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "private_network.0.private_network_id", "scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.ip"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_rdb_read_replica.replica", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "direct_access.#", "0"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.#", "1"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.0.enable_ipam", "false"),
				),
			},
		},
	})
}

func TestAccScalewayRdbReadReplica_MultipleEndpoints(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := testAccCheckScalewayRdbEngineGetLatestVersion(tt, postgreSQLEngineName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-multiple-endpoints"
						node_type = "db-dev-s"
						engine = %q
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "multiple-endpoints" ]
					}

					resource "scaleway_vpc_private_network" "pn" {}

					resource "scaleway_rdb_read_replica" "replica" {
  						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
							service_ip         = "10.12.1.0/20"
						}
						direct_access {}
					}`, latestEngineVersion),
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

func TestAccScalewayRdbReadReplica_DifferentZone(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	readReplicaID := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_private_network" "different_zone" {
						name = "test-rdb-rr-different-zone"
					}

					resource "scaleway_rdb_instance" "different_zone" {
						name = "test-rdb-rr-different-zone"
						node_type = "db-dev-s"
						engine = "PostgreSQL-14"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "different-zone" ]
					}

					resource "scaleway_rdb_read_replica" "different_zone" {
  						instance_id = scaleway_rdb_instance.different_zone.id
						region = scaleway_rdb_instance.different_zone.region
						private_network {
							private_network_id = scaleway_vpc_private_network.different_zone.id
							enable_ipam = true
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.different_zone"),
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.different_zone"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.different_zone", "instance_id", "scaleway_rdb_instance.different_zone", "id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.different_zone", "same_zone", "true"),
					acctest.TestAccCheckScalewayResourceIDPersisted("scaleway_rdb_read_replica.different_zone", &readReplicaID),
				),
			},
			{
				Config: `
					resource "scaleway_vpc_private_network" "different_zone" {
						name = "test-rdb-rr-different-zone"
					}

					resource "scaleway_rdb_instance" "different_zone" {
						name = "test-rdb-rr-different-zone"
						node_type = "db-dev-s"
						engine = "PostgreSQL-14"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "different-zone" ]
					}

					resource "scaleway_rdb_read_replica" "different_zone" {
  						instance_id = scaleway_rdb_instance.different_zone.id
						region = scaleway_rdb_instance.different_zone.region
						same_zone = true
						private_network {
							private_network_id = scaleway_vpc_private_network.different_zone.id
							enable_ipam = true
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.different_zone"),
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.different_zone"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.different_zone", "instance_id", "scaleway_rdb_instance.different_zone", "id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.different_zone", "same_zone", "true"),
					acctest.TestAccCheckScalewayResourceIDPersisted("scaleway_rdb_read_replica.different_zone", &readReplicaID),
				),
			},
			{
				Config: `
					resource "scaleway_vpc_private_network" "different_zone" {
						name = "test-rdb-rr-different-zone"
					}

					resource "scaleway_rdb_instance" "different_zone" {
						name = "test-rdb-rr-different-zone"
						node_type = "db-dev-s"
						engine = "PostgreSQL-14"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "different-zone" ]
					}

					resource "scaleway_rdb_read_replica" "different_zone" {
  						instance_id = scaleway_rdb_instance.different_zone.id
						region = scaleway_rdb_instance.different_zone.region
						same_zone = false
						private_network {
							private_network_id = scaleway_vpc_private_network.different_zone.id
							enable_ipam = true
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.different_zone"),
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.different_zone"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.different_zone", "instance_id", "scaleway_rdb_instance.different_zone", "id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.different_zone", "same_zone", "false"),
					acctest.TestAccCheckScalewayResourceIDChanged("scaleway_rdb_read_replica.different_zone", &readReplicaID),
				),
			},
		},
	})
}

func TestAccScalewayRdbReadReplica_WithInstanceAlsoInPrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayRdbReadReplicaDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_private_network" "pn1" {
						name = "test-rdb-rr-instance-in-pn1"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						name = "test-rdb-rr-instance-in-pn2"
					}

					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-instance-in-pn"
						node_type = "db-dev-s"
						engine = "PostgreSQL-15"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "instance-also-in-pn" ]
						private_network {
							pn_id = scaleway_vpc_private_network.pn1.id
							enable_ipam = true
						}
					}

					resource "scaleway_rdb_read_replica" "replica" {
						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn1.id
							enable_ipam = true
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "direct_access.#", "0"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "private_network.0.private_network_id", "scaleway_vpc_private_network.pn1", "id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.0.enable_ipam", "true"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_instance.instance", "private_network.0.pn_id", "scaleway_vpc_private_network.pn1", "id"),
					resource.TestCheckResourceAttr("scaleway_rdb_instance.instance", "private_network.0.enable_ipam", "true"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc_private_network" "pn1" {
						name = "test-rdb-rr-instance-in-pn1"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						name = "test-rdb-rr-instance-in-pn2"
					}
			
					resource scaleway_rdb_instance instance {
						name = "test-rdb-rr-instance-in-pn"
						node_type = "db-dev-s"
						engine = "PostgreSQL-15"
						is_ha_cluster = false
						disable_backup = true
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						tags = [ "terraform-test", "scaleway_rdb_read_replica", "instance-also-in-pn" ]
						private_network {
							pn_id = scaleway_vpc_private_network.pn2.id
							enable_ipam = true
						}
					}
			
					resource "scaleway_rdb_read_replica" "replica" {
						instance_id = scaleway_rdb_instance.instance.id
						private_network {
							private_network_id = scaleway_vpc_private_network.pn2.id
							enable_ipam = true
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbReadReplicaExists(tt, "scaleway_rdb_read_replica.replica"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "instance_id", "scaleway_rdb_instance.instance", "id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "direct_access.#", "0"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_rdb_read_replica.replica", "private_network.0.private_network_id", "scaleway_vpc_private_network.pn2", "id"),
					resource.TestCheckResourceAttr("scaleway_rdb_read_replica.replica", "private_network.0.enable_ipam", "true"),
				),
			},
		},
	})
}

func testAccCheckRdbReadReplicaExists(tt *acctest.TestTools, readReplica string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		readReplicaResource, ok := state.RootModule().Resources[readReplica]
		if !ok {
			return fmt.Errorf("resource not found: %s", readReplica)
		}

		rdbAPI, region, ID, err := scaleway.RdbAPIWithRegionAndID(tt.Meta, readReplicaResource.Primary.ID)
		if err != nil {
			return err
		}

		_, err = rdbAPI.GetReadReplica(&rdb.GetReadReplicaRequest{
			Region:        region,
			ReadReplicaID: ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayRdbReadReplicaDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_rdb_read_replica" {
				continue
			}

			rdbAPI, region, ID, err := scaleway.RdbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
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
			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
