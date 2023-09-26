package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
)

func TestAccScalewayDocumentDBInstanceEndpoint_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayDocumentDBInstanceEndpointDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_document_db_instance" "main" {
				  name              = "test-documentdb-instance-endpoint-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  is_ha_cluster     = false
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  tags              = ["terraform-test", "scaleway_document_db_instance", "minimal"]
				  volume_size_in_gb = 20
				  telemetry_enabled = false
				}

				resource "scaleway_vpc_private_network" "pn" {
				  name = "my_private_network"
				}

				resource "scaleway_document_db_instance_private_network_endpoint" "main" {
				  instance_id = scaleway_document_db_instance.main.id
					ip_net = "172.16.32.3/22"
					private_network_id     = scaleway_vpc_private_network.pn.id
				  depends_on = [scaleway_vpc_private_network.pn]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "scaleway_document_db_instance_endpoint.main"),
					testCheckResourceAttrUUID("scaleway_document_db_instance_endpoint.main", "id"),
					resource.TestCheckResourceAttr("scaleway_document_db_instance.main", "name", "test-documentdb-instance-endpoint-basic"),
					resource.TestCheckResourceAttr(
						"scaleway_document_db_instance_endpoint.main", "ip_net", "172.16.32.3/22"),
				),
			},
			{
				Config: `
				resource "scaleway_document_db_instance" "main" {
				  name              = "test-documentdb-instance-endpoint-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  is_ha_cluster     = false
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  tags              = ["terraform-test", "scaleway_document_db_instance", "minimal"]
				  volume_size_in_gb = 20
				  telemetry_enabled = false
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

				resource "scaleway_document_db_instance_private_network_endpoint" "main" {
				  instance_id = scaleway_document_db_instance.main.id
					ip_net = "172.16.32.3/22"
					private_network_id     = scaleway_vpc_private_network.pn.id
				  depends_on = [scaleway_vpc_private_network.pn02, scaleway_vpc.vpc]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "scaleway_document_db_instance_endpoint.main"),
					testCheckResourceAttrUUID("scaleway_document_db_instance_endpoint.main", "id"),
					resource.TestCheckResourceAttr("scaleway_document_db_instance.main", "name", "test-documentdb-instance-endpoint-basic"),
					resource.TestCheckResourceAttr(
						"scaleway_document_db_instance_endpoint.main", "ip_net", "172.16.32.3/22"),
				),
			},
			{
				Config: `
				resource "scaleway_document_db_instance" "main" {
				  name              = "test-documentdb-instance-endpoint-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  is_ha_cluster     = false
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  tags              = ["terraform-test", "scaleway_document_db_instance", "minimal"]
				  volume_size_in_gb = 20
				  telemetry_enabled = false
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

				resource "scaleway_document_db_instance_private_network_endpoint" "main" {
				  instance_id = scaleway_document_db_instance.main.id
					ip_net = "172.16.64.4/22"
					private_network_id     = scaleway_vpc_private_network.pn02.id
				  depends_on = [scaleway_vpc_private_network.pn02, scaleway_vpc.vpc]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "scaleway_document_db_instance_endpoint.main"),
					testCheckResourceAttrUUID("scaleway_document_db_instance_endpoint.main", "id"),
					resource.TestCheckResourceAttr("scaleway_document_db_instance.main", "name", "test-documentdb-instance-endpoint-basic"),
					resource.TestCheckResourceAttr(
						"scaleway_document_db_instance_endpoint.main", "ip_net", "172.16.64.4/22"),
				),
			},
			{
				Config: `
				resource "scaleway_document_db_instance" "main" {
				  name              = "test-documentdb-instance-endpoint-basic"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  is_ha_cluster     = false
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  tags              = ["terraform-test", "scaleway_document_db_instance", "minimal"]
				  volume_size_in_gb = 20
				  telemetry_enabled = false
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

				resource "scaleway_document_db_instance_private_network_endpoint" "main" {
				  instance_id = scaleway_document_db_instance.main.id
					ip_net = "172.16.64.4/22"
					private_network_id     = scaleway_vpc_private_network.pn02.id
				  depends_on = [scaleway_vpc_private_network.pn02, scaleway_vpc.vpc]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "scaleway_document_db_instance_endpoint.main"),
					testCheckResourceAttrUUID("scaleway_document_db_instance_private_network_endpoint.main", "id"),
					resource.TestCheckResourceAttr("scaleway_document_db_instance.main", "name", "test-documentdb-instance-endpoint-basic"),
					resource.TestCheckResourceAttr(
						"scaleway_document_db_instance_private_network_endpoint.main", "ip_net", "172.16.64.4/22"),
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

func TestAccScalewayDocumentDBInstanceEndpoint_Migration(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayDocumentDBInstanceEndpointDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_document_db_instance" "main" {
				  name              = "test-documentdb-instance-endpoint-migration"
				  node_type         = "docdb-play2-pico"
				  engine            = "FerretDB-1"
				  is_ha_cluster     = false
				  user_name         = "my_initial_user"
				  password          = "thiZ_is_v&ry_s3cret"
				  tags              = ["terraform-test", "scaleway_document_db_instance_migration", "minimal"]
				  volume_size_in_gb = 20
				  telemetry_enabled = false
				}

				resource scaleway_vpc vpc {
					name = "vpcDocumentDB"
				}

				resource "scaleway_vpc_private_network" "pn" {
					ipv4_subnet {
						subnet = "10.10.64.0/22"
					}
					vpc_id = scaleway_vpc.vpc.id
				}	

				resource "scaleway_document_db_instance_private_network_endpoint" "main" {
				  instance_id = scaleway_document_db_instance.main.id
					ip_net = "10.10.64.4/22"
					private_network_ide      = scaleway_vpc_private_network.pn.id
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayDocumentDBInstanceEndpointExists(tt, "scaleway_document_db_instance_endpoint.main"),
					testCheckResourceAttrUUID("scaleway_document_db_instance_endpoint.main", "id"),
					resource.TestCheckResourceAttr(
						"scaleway_document_db_instance_endpoint.main", "ip_net", "10.10.64.4/22"),
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

func testAccCheckScalewayDocumentDBInstanceEndpointDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_document_db_instance_endpoint" {
				continue
			}

			api, region, id, err := documentDBAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteEndpoint(&documentdb.DeleteEndpointRequest{
				Region:     region,
				EndpointID: id,
			})

			if err == nil {
				return fmt.Errorf("documentdb documentdb instance endpoint (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}

func testAccCheckScalewayDocumentDBInstanceEndpointExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := documentDBAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
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
