package scaleway

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	redis "github.com/scaleway/scaleway-sdk-go/api/redis/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_redis_cluster", &resource.Sweeper{
		Name: "scaleway_redis_cluster",
		F:    testSweepRedisCluster,
	})
}

func testSweepRedisCluster(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		redisAPI := redis.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the redis cluster in (%s)", zone)
		listClusters, err := redisAPI.ListClusters(&redis.ListClustersRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing redis clusters in (%s) in sweeper: %w", zone, err)
		}

		for _, cluster := range listClusters.Clusters {
			_, err := redisAPI.DeleteCluster(&redis.DeleteClusterRequest{
				Zone:      zone,
				ClusterID: cluster.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting redis cluster in sweeper: %w", err)
			}
		}

		return nil
	})
}

func TestAccScalewayRedisCluster_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRedisClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_basic"
    					version = "6.2.6"
    					node_type = "RED1-XS"
    					user_name = "my_initial_user"
    					password = "thiZ_is_v&ry_s3cret"
						tags = [ "test1" ]
						cluster_size = 1
						tls_enabled = "true"
						zone = "fr-par-2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_basic"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tags.0", "test1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tls_enabled", "true"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "zone", "fr-par-2"),
				),
			},
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_basic_edit"
    					version = "6.2.6"
    					node_type = "RED1-XS"
    					user_name = "new_user"
    					password = "thiZ_is_A_n3w_passw0rd"
						tags = [ "test1", "other_tag" ]
						cluster_size = 1
						tls_enabled = "true"
						zone = "fr-par-2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_basic_edit"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "new_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_A_n3w_passw0rd"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tags.0", "test1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tags.1", "other_tag"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tls_enabled", "true"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "zone", "fr-par-2"),
				),
			},
		},
	})
}

func TestAccScalewayRedisCluster_Migrate(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRedisClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_basic"
    					version = "6.2.6"
    					node_type = "RED1-XS"
    					user_name = "my_initial_user"
    					password = "thiZ_is_v&ry_s3cret"
						tags = [ "test1" ]
						cluster_size = 1
						tls_enabled = "true"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_basic"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tags.0", "test1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tls_enabled", "true"),
				),
			},
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_basic"
    					version = "6.2.6"
    					node_type = "RED1-S"
    					user_name = "my_initial_user"
    					password = "thiZ_is_v&ry_s3cret"
						tags = [ "test1" ]
						cluster_size = 1
						tls_enabled = "true"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_basic"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-S"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tags.0", "test1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tls_enabled", "true"),
				),
			},
		},
	})
}

func TestAccScalewayRedisCluster_ACL(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRedisClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_acl"
    					version = "6.2.6"
    					node_type = "RED1-XS"
    					user_name = "my_initial_user"
    					password = "thiZ_is_v&ry_s3cret"
						acl {
							ip = "0.0.0.0/0"
							description = "An acl description"
						}
						acl {
							ip = "192.168.10.0/24"
							description = "A second acl description"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_acl"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_redis_cluster.main", "acl.*", map[string]string{
						"ip":          "0.0.0.0/0",
						"description": "An acl description",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_redis_cluster.main", "acl.*", map[string]string{
						"ip":          "192.168.10.0/24",
						"description": "A second acl description",
					}),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "acl.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "acl.1.id"),
				),
			},
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_acl"
    					version = "6.2.6"
    					node_type = "RED1-XS"
    					user_name = "my_initial_user"
    					password = "thiZ_is_v&ry_s3cret"
						acl {
							ip = "192.168.11.0/24"
							description = "Another acl description"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_acl"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "acl.0.ip", "192.168.11.0/24"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "acl.0.description", "Another acl description"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "acl.0.id"),
				),
			},
		},
	})
}

func TestAccScalewayRedisCluster_Settings(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRedisClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_settings"
    					version = "6.2.6"
    					node_type = "RED1-XS"
    					user_name = "my_initial_user"
    					password = "thiZ_is_v&ry_s3cret"
						settings = {
							"tcp-keepalive" = "150"
							"maxclients" = "5000"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_settings"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "settings.tcp-keepalive", "150"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "settings.maxclients", "5000"),
				),
			},
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_settings"
    					version = "6.2.6"
    					node_type = "RED1-XS"
    					user_name = "my_initial_user"
    					password = "thiZ_is_v&ry_s3cret"
						settings = {
							"maxclients" = "2000"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_settings"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "settings.maxclients", "2000"),
				),
			},
		},
	})
}

func TestAccScalewayRedisCluster_Endpoints_Standalone(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRedisClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				// Step 1: First we define a single private network
				Config: `
					resource "scaleway_vpc_private_network" "pn" {
						name = "private-network"
					}
					resource "scaleway_redis_cluster" "main" {
						name =			"test_redis_endpoints"
						version = 		"6.2.6"
						node_type = 	"RED1-XS"
						user_name = 	"my_initial_user"
						password = 		"thiZ_is_v&ry_s3cret"
						cluster_size = 	1
						private_network {
							id = "${scaleway_vpc_private_network.pn.id}"
							service_ips = [
								"10.12.1.0/20",
							]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.pn"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_endpoints"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "private_network.0.service_ips.0", "10.12.1.0/20"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.0.id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_redis_cluster.main", "private_network.0.id", "scaleway_vpc_private_network.pn", "id"),
				),
			},
			{
				// Step 2: Then we add another one
				Config: `
					resource "scaleway_vpc_private_network" "pn" {
						name = "private-network"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						name = "private-network-2"
					}
					resource "scaleway_redis_cluster" "main" {
						name =			"test_redis_endpoints"
						version = 		"6.2.6"
						node_type = 	"RED1-XS"
						user_name = 	"my_initial_user"
						password = 		"thiZ_is_v&ry_s3cret"
						cluster_size = 	1
						private_network {
							id = "${scaleway_vpc_private_network.pn.id}"
							service_ips = [
								"10.12.1.0/20",
							]
						}
						private_network {
							id = "${scaleway_vpc_private_network.pn2.id}"
							service_ips = [
								"192.168.1.0/20",
							]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.pn"),
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.pn2"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_endpoints"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.1.id"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.1.endpoint_id"),
					testAccCheckScalewayRedisPrivateNetworksIpsAreEither("scaleway_redis_cluster.main", "10.12.1.0/20", "192.168.1.0/20"),
					testAccCheckScalewayRedisPrivateNetworksIdsAreEither("scaleway_redis_cluster.main", "scaleway_vpc_private_network.pn", "scaleway_vpc_private_network.pn2"),
				),
			},
			{
				// Step 3: Then we modify the first one and remove the second one
				Config: `
					resource "scaleway_vpc_private_network" "pn" {
						name = "private-network"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						name = "private-network-2"
					}
					resource "scaleway_redis_cluster" "main" {
						name =			"test_redis_endpoints"
						version = 		"6.2.6"
						node_type = 	"RED1-XS"
						user_name = 	"my_initial_user"
						password = 		"thiZ_is_v&ry_s3cret"
						cluster_size = 	1
						private_network {
							id = "${scaleway_vpc_private_network.pn.id}"
							service_ips = [
								"10.13.1.0/20",
							]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.pn"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_endpoints"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "private_network.0.service_ips.0", "10.13.1.0/20"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.0.endpoint_id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_redis_cluster.main", "private_network.0.id", "scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckNoResourceAttr("scaleway_redis_cluster.main", "private_network.1.service_ips.0"),
					resource.TestCheckNoResourceAttr("scaleway_redis_cluster.main", "private_network.1.id"),
					resource.TestCheckNoResourceAttr("scaleway_redis_cluster.main", "private_network.1.endpoint_id"),
				),
			},
			{
				// Step 4: And finally we remove the private network to check that we still have a public network
				Config: `
					resource "scaleway_vpc_private_network" "pn" {
						name = "private-network"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						name = "private-network-2"
					}
					resource "scaleway_redis_cluster" "main" {
						name = "test_redis_endpoints"
						version = "6.2.6"
						node_type = "RED1-XS"
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"
						cluster_size = 1
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_endpoints"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckNoResourceAttr("scaleway_redis_cluster.main", "private_network.0.id"),
					resource.TestCheckNoResourceAttr("scaleway_redis_cluster.main", "private_network.0.port"),
					resource.TestCheckNoResourceAttr("scaleway_redis_cluster.main", "private_network.0.ips.#"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "public_network.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "public_network.0.port"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "public_network.0.ips.#"),
				),
			},
			{
				// Step 5: Extra step just to be sure that the cluster is deleted before the Private Networks
				Config: `
					resource "scaleway_vpc_private_network" "pn" {
						name = "private-network"
					}
					resource "scaleway_vpc_private_network" "pn2" {
						name = "private-network-2"
					}
				`,
			},
		},
	})
}

func TestAccScalewayRedisCluster_Endpoints_ClusterMode(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRedisClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				// Step 1: We define a single private network
				Config: `
					resource "scaleway_vpc_private_network" "pn" {
						name = "private-network"
					}
					resource "scaleway_redis_cluster" "main" {
						name =			"test_redis_endpoints_cluster_mode"
						version = 		"6.2.6"
						node_type = 	"RED1-XS"
						user_name = 	"my_initial_user"
						password = 		"thiZ_is_v&ry_s3cret"
						cluster_size = 	3
						private_network {
							id = "${scaleway_vpc_private_network.pn.id}"
							service_ips = [
								"10.12.1.10/24",
								"10.12.1.11/24",
								"10.12.1.12/24",
							]
						}
						depends_on = [
							scaleway_vpc_private_network.pn,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.pn"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_endpoints_cluster_mode"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "3"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "private_network.0.service_ips.0", "10.12.1.10/24"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "private_network.0.service_ips.1", "10.12.1.11/24"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "private_network.0.service_ips.2", "10.12.1.12/24"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.0.endpoint_id"),
					resource.TestCheckResourceAttrSet("scaleway_redis_cluster.main", "private_network.0.id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_redis_cluster.main", "private_network.0.id", "scaleway_vpc_private_network.pn", "id"),
				),
			},
			{
				// Step 2: We delete the cluster, but keep the private network to be sure it's not deleted before
				Config: `
					resource "scaleway_vpc_private_network" "pn" {
						name = "private-network"
					}
				`,
			},
		},
	})
}

func TestAccScalewayRedisCluster_Certificate(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRedisClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_certificate"
    					version = "6.2.6"
    					node_type = "RED1-XS"
    					user_name = "my_initial_user"
    					password = "thiZ_is_v&ry_s3cret"
						tags = [ "test1" ]
						cluster_size = 1
						tls_enabled = "true"
						zone = "fr-par-2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_certificate"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tags.0", "test1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tls_enabled", "true"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "zone", "fr-par-2"),
					testAccCheckScalewayRedisCertificateIsValid("scaleway_redis_cluster.main"),
				),
			},
		},
	})
}

func TestAccScalewayRedisCluster_NoCertificate(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRedisClusterDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_redis_cluster" "main" {
    					name = "test_redis_no_certificate"
    					version = "6.2.6"
    					node_type = "RED1-XS"
    					user_name = "my_initial_user"
    					password = "thiZ_is_v&ry_s3cret"
						tags = [ "test1" ]
						cluster_size = 1
						tls_enabled = "false"
						zone = "fr-par-2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRedisExists(tt, "scaleway_redis_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "name", "test_redis_no_certificate"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "version", "6.2.6"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "node_type", "RED1-XS"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "user_name", "my_initial_user"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "password", "thiZ_is_v&ry_s3cret"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tags.0", "test1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "cluster_size", "1"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "tls_enabled", "false"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "zone", "fr-par-2"),
					resource.TestCheckResourceAttr("scaleway_redis_cluster.main", "certificate", ""),
				),
			},
		},
	})
}

func testAccCheckScalewayRedisClusterDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_redis_cluster" {
				continue
			}

			redisAPI, zone, ID, err := redisAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = redisAPI.GetCluster(&redis.GetClusterRequest{
				ClusterID: ID,
				Zone:      zone,
			})

			if err == nil {
				return fmt.Errorf("cluster (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}
		return nil
	}
}

func testAccCheckScalewayRedisExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		redisAPI, zone, ID, err := redisAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = redisAPI.GetCluster(&redis.GetClusterRequest{
			ClusterID: ID,
			Zone:      zone,
		})

		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckScalewayRedisPrivateNetworksIpsAreEither(name string, possibilities ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}
		actualIPs := []string(nil)
		for i := range possibilities {
			actualIPs = append(actualIPs, rs.Primary.Attributes[fmt.Sprintf("private_network.%d.service_ips.0", i)])
		}
		for _, ip := range actualIPs {
			for i := range possibilities {
				if possibilities[i] == ip {
					possibilities[i] = "ip found"
				}
			}
		}
		for _, p := range possibilities {
			if p != "ip found" {
				return fmt.Errorf("no attribute private_network.*.service_ips.0 was found with value %v", p)
			}
		}
		return nil
	}
}

func testAccCheckScalewayRedisPrivateNetworksIdsAreEither(name string, possibilities ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}
		for i, possibility := range possibilities {
			rs, ok := state.RootModule().Resources[possibility]
			if ok {
				possibilities[i] = rs.Primary.ID
			}
		}
		actualIDs := []string(nil)
		for i := range possibilities {
			toLookFor := fmt.Sprintf("private_network.%d.id", i)
			id := rs.Primary.Attributes[toLookFor]
			actualIDs = append(actualIDs, id)
		}
		for _, id := range actualIDs {
			for i := range possibilities {
				if possibilities[i] == id {
					possibilities[i] = "id found"
				}
			}
		}
		for _, p := range possibilities {
			if p != "id found" {
				return fmt.Errorf("no attribute private_network.*.id was found with value %v", p)
			}
		}
		return nil
	}
}

func testAccCheckScalewayRedisCertificateIsValid(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}
		pemCert, hasCert := rs.Primary.Attributes["certificate"]
		if !hasCert {
			return fmt.Errorf("could not find certificate in schema")
		}
		cert, _ := pem.Decode([]byte(pemCert))
		_, err := x509.ParseCertificate(cert.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}
		return nil
	}
}
