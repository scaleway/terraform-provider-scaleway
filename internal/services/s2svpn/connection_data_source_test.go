package s2svpn_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceConnection_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-test-vpc-connection-ds"
					}

					resource "scaleway_vpc_private_network" "main" {
						vpc_id = scaleway_vpc.main.id
						ipv4_subnet {
							subnet = "10.0.0.0/24"
						}
					}

					resource "scaleway_instance_ip" "customer_ip" {}

					resource "scaleway_s2s_vpn_gateway" "main" {
						name              = "tf-test-vpn-gateway-conn-ds"
						gateway_type      = "VGW-S"
						private_network_id = scaleway_vpc_private_network.main.id
						region            = "fr-par"
						zone              = "fr-par-1"
					}

					resource "scaleway_s2s_vpn_customer_gateway" "main" {
						name        = "tf-test-customer-gateway-conn-ds"
						asn         = 65000
						ipv4_public = scaleway_instance_ip.customer_ip.address
						region     = "fr-par"
					}

					resource "scaleway_s2s_vpn_routing_policy" "main" {
						name              = "tf-test-routing-policy-conn-ds"
						prefix_filter_in  = ["10.0.1.0/24"]
						prefix_filter_out = ["10.0.0.0/24"]
						region            = "fr-par"					
					}

					resource "scaleway_s2s_vpn_connection" "main" {
						name                = "tf-test-connection-ds"
						vpn_gateway_id      = scaleway_s2s_vpn_gateway.main.id
						customer_gateway_id = scaleway_s2s_vpn_customer_gateway.main.id
					
						bgp_config_ipv4 {
							routing_policy_id = scaleway_s2s_vpn_routing_policy.main.id
							private_ip        = "169.254.0.1/30"
							peer_private_ip   = "169.254.0.2/30"
						}

						ikev2_ciphers {
							encryption = "aes256"
							integrity  = "sha256"
							dh_group   = "modp2048"
						}

						esp_ciphers {
							encryption = "aes256"
							integrity  = "sha256"
							dh_group   = "modp2048"
						}
					}
				`,
			},
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-test-vpc-connection-ds"
					}

					resource "scaleway_vpc_private_network" "main" {
						vpc_id = scaleway_vpc.main.id
						ipv4_subnet {
							subnet = "10.0.0.0/24"
						}
					}

					resource "scaleway_instance_ip" "customer_ip" {}

					resource "scaleway_s2s_vpn_gateway" "main" {
						name              = "tf-test-vpn-gateway-conn-ds"
						gateway_type      = "VGW-S"
						private_network_id = scaleway_vpc_private_network.main.id
						region            = "fr-par"
						zone              = "fr-par-1"
					}

					resource "scaleway_s2s_vpn_customer_gateway" "main" {
						name        = "tf-test-customer-gateway-conn-ds"
						asn         = 65000
						ipv4_public = scaleway_instance_ip.customer_ip.address
						region     = "fr-par"
					}

					resource "scaleway_s2s_vpn_routing_policy" "main" {
						name             = "tf-test-routing-policy-conn-ds"
						prefix_filter_in  = ["10.0.1.0/24"]
						prefix_filter_out = ["10.0.0.0/24"]
						region          = "fr-par"					
					}

					resource "scaleway_s2s_vpn_connection" "main" {
						name                = "tf-test-connection-ds"
						vpn_gateway_id      = scaleway_s2s_vpn_gateway.main.id
						customer_gateway_id = scaleway_s2s_vpn_customer_gateway.main.id
						
						bgp_config_ipv4 {
							routing_policy_id = scaleway_s2s_vpn_routing_policy.main.id
							private_ip        = "169.254.0.1/30"
							peer_private_ip   = "169.254.0.2/30"
						}

						ikev2_ciphers {
							encryption = "aes256"
							integrity  = "sha256"
							dh_group   = "modp2048"
						}

						esp_ciphers {
							encryption = "aes256"
							integrity  = "sha256"
							dh_group   = "modp2048"
						}
					}

					data "scaleway_s2s_vpn_connection" "by_name" {
						name = scaleway_s2s_vpn_connection.main.name
					}

					data "scaleway_s2s_vpn_connection" "by_id" {
						connection_id = scaleway_s2s_vpn_connection.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(tt, "scaleway_s2s_vpn_connection.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_connection.by_name", "name",
						"scaleway_s2s_vpn_connection.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_connection.by_name", "vpn_gateway_id",
						"scaleway_s2s_vpn_connection.main", "vpn_gateway_id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_connection.by_id", "connection_id",
						"scaleway_s2s_vpn_connection.main", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_connection.by_id", "name",
						"scaleway_s2s_vpn_connection.main", "name"),
				),
			},
		},
	})
}
