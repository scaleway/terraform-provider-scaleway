package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func TestAccScalewayDocumentDBPrivateNetworkEndpoint_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayDocumentDBInstanceEndpointDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_documentdb_instance" "main" {
				  name              = "test-documentdb-instance-endpoint-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 20
				}
				
				resource "scaleway_vpc_private_network" "pn" {
				  name = "my_private_network"
				}
				
				resource "scaleway_documentdb_private_network_endpoint" "main" {
				  instance_id        = scaleway_documentdb_instance.main.id
				  ip_net             = "172.16.32.3/22"
				  private_network_id = scaleway_vpc_private_network.pn.id
				  depends_on         = [scaleway_vpc_private_network.pn]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "scaleway_documentdb_private_network_endpoint.main"),
					testCheckResourceAttrUUID("scaleway_documentdb_private_network_endpoint.main", "id"),
					resource.TestCheckResourceAttr("scaleway_documentdb_instance.main", "name", "test-documentdb-instance-endpoint-basic"),
					resource.TestCheckResourceAttr(
						"scaleway_documentdb_private_network_endpoint.main", "ip_net", "172.16.32.3/22"),
				),
			},
			{
				Config: `
				resource "scaleway_documentdb_instance" "main" {
				  name              = "test-documentdb-instance-endpoint-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 20
				}

				resource "scaleway_vpc_private_network" "pn" {
				  name = "my_private_network"
				}

				resource "scaleway_vpc" "vpc" {
				  name = "my vpc"
				}

				// Creation to the new private network with new subnet
				resource "scaleway_vpc_private_network" "pn02" {
				  ipv4_subnet {
					subnet = "172.16.64.0/22"
				  }
				  vpc_id = scaleway_vpc.vpc.id
				}

				resource "scaleway_documentdb_private_network_endpoint" "main" {
				  instance_id        = scaleway_documentdb_instance.main.id
				  ip_net             = "172.16.32.3/22"
				  private_network_id = scaleway_vpc_private_network.pn.id
				  depends_on         = [scaleway_vpc_private_network.pn02, scaleway_vpc.vpc]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "scaleway_documentdb_private_network_endpoint.main"),
					testCheckResourceAttrUUID("scaleway_documentdb_private_network_endpoint.main", "id"),
					resource.TestCheckResourceAttr("scaleway_documentdb_instance.main", "name", "test-documentdb-instance-endpoint-basic"),
					resource.TestCheckResourceAttr(
						"scaleway_documentdb_private_network_endpoint.main", "ip_net", "172.16.32.3/22"),
				),
			},
			{
				Config: `
				resource "scaleway_documentdb_instance" "main" {
				  name              = "test-documentdb-instance-endpoint-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  volume_size_in_gb = 20
				}

				resource "scaleway_vpc_private_network" "pn" {
				  name = "my_private_network"
				}

				resource "scaleway_vpc" "vpc" {
				  name = "my vpc"
				}

				resource "scaleway_vpc_private_network" "pn02" {
				  ipv4_subnet {
					subnet = "172.16.64.0/22"
				  }
				  vpc_id = scaleway_vpc.vpc.id
				}

				// Replace the ip on the new private network
				resource "scaleway_documentdb_private_network_endpoint" "main" {
				  instance_id        = scaleway_documentdb_instance.main.id
				  ip_net             = "172.16.64.4/22"
				  private_network_id = scaleway_vpc_private_network.pn02.id
				  depends_on         = [scaleway_vpc_private_network.pn02, scaleway_vpc.vpc]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "scaleway_documentdb_private_network_endpoint.main"),
					testCheckResourceAttrUUID("scaleway_documentdb_private_network_endpoint.main", "id"),
					resource.TestCheckResourceAttr("scaleway_documentdb_instance.main", "name", "test-documentdb-instance-endpoint-basic"),
					resource.TestCheckResourceAttr(
						"scaleway_documentdb_private_network_endpoint.main", "ip_net", "172.16.64.4/22"),
				),
			},
			{
				Config: `
				resource "scaleway_vpc_private_network" "pn" {
				  name = "my_private_network"
				}

				resource "scaleway_vpc" "vpc" {
				  name = "my vpc"
				}

				resource "scaleway_vpc_private_network" "pn02" {
				  ipv4_subnet {
					subnet = "172.16.64.0/22"
				  }
				  vpc_id = scaleway_vpc.vpc.id
				}
				`,
			},
		},
	})
}

func TestAccScalewayDocumentDBPrivateNetworkEndpoint_Migration(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayDocumentDBInstanceEndpointDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_documentdb_instance" "main" {
				  name              = "test-documentdb-instance-endpoint-migration"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  tags              = ["terraform-test", "scaleway_documentdb_instance_migration", "minimal"]
				  volume_size_in_gb = 20
				}

				resource "scaleway_vpc" "vpc" {
				  name = "vpcDocumentDB"
				}

				resource "scaleway_vpc_private_network" "pn" {
				  ipv4_subnet {
					subnet = "10.10.64.0/22"
				  }
				  vpc_id = scaleway_vpc.vpc.id
				}

				resource "scaleway_documentdb_private_network_endpoint" "main" {
				  instance_id         = scaleway_documentdb_instance.main.id
				  ip_net              = "10.10.64.4/22"
				  private_network_id = scaleway_vpc_private_network.pn.id
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "scaleway_documentdb_private_network_endpoint.main"),
					testCheckResourceAttrUUID("scaleway_documentdb_private_network_endpoint.main", "id"),
					resource.TestCheckResourceAttr(
						"scaleway_documentdb_private_network_endpoint.main", "ip_net", "10.10.64.4/22"),
				),
			},
			{
				Config: `
				resource scaleway_vpc vpc {
					name = "vpc"
				}

				resource "scaleway_vpc_private_network" "pn" {
					ipv4_subnet {
						subnet = "10.10.64.0/22"
					}
					vpc_id = scaleway_vpc.vpc.id
				}
				`,
			},
		},
	})
}

func testAccCheckScalewayDocumentDBInstanceEndpointDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_documentdb_private_network_endpoint" {
				continue
			}

			api, region, id, err := scaleway.DocumentDBAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteEndpoint(&documentdb.DeleteEndpointRequest{
				Region:     region,
				EndpointID: id,
			})

			if err == nil {
				return fmt.Errorf("documentdb instance endpoint (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}

func testAccCheckScalewayDocumentDBInstanceEndpointExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := scaleway.DocumentDBAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetEndpoint(&documentdb.GetEndpointRequest{
			EndpointID: id,
			Region:     region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
