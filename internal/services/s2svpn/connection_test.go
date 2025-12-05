package s2svpn_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/s2svpn"
)

func TestAccConnection_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckConnectionDestroy(tt),
			testAccCheckVPNGatewayDestroy(tt),
			testAccCheckCustomerGatewayDestroy(tt),
			testAccCheckRoutingPolicyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-test-vpc-connection"
					}

					resource "scaleway_vpc_private_network" "main" {
						vpc_id = scaleway_vpc.main.id
						ipv4_subnet {
							subnet = "10.0.0.0/24"
						}
					}

					resource "scaleway_instance_ip" "customer_ip" {}

					resource "scaleway_s2s_vpn_gateway" "main" {
						name              = "tf-test-vpn-gateway-connection"
						gateway_type      = "VGW-S"
						private_network_id = scaleway_vpc_private_network.main.id
						region            = "fr-par"
						zone              = "fr-par-1"
					}

					resource "scaleway_s2s_vpn_customer_gateway" "main" {
						name       = "tf-test-customer-gateway-connection"
						ipv4_public = scaleway_instance_ip.customer_ip.address
						asn        = 65000
						region     = "fr-par"
					}

					resource "scaleway_s2s_vpn_routing_policy" "main" {
						name            = "tf-test-routing-policy-connection"
						prefix_filter_in  = ["10.0.1.0/24"]
						prefix_filter_out = ["10.0.0.0/24"]
						region          = "fr-par"
					}

					resource "scaleway_s2s_vpn_connection" "main" {
						name                     = "tf-test-connection"
						vpn_gateway_id           = scaleway_s2s_vpn_gateway.main.id
						customer_gateway_id      = scaleway_s2s_vpn_customer_gateway.main.id
						initiation_policy        = "customer_gateway"
						enable_route_propagation = true
						region                   = "fr-par"

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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(tt, "scaleway_s2s_vpn_connection.main"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "name", "tf-test-connection"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "initiation_policy", "customer_gateway"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "enable_route_propagation", "true"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_connection.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_connection.main", "status"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_connection.main", "secret_id"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_connection.main", "secret_version"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "bgp_config_ipv4.0.private_ip", "169.254.0.1/30"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "bgp_config_ipv4.0.peer_private_ip", "169.254.0.2/30"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "ikev2_ciphers.0.encryption", "aes256"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "ikev2_ciphers.0.integrity", "sha256"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "ikev2_ciphers.0.dh_group", "modp2048"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "esp_ciphers.0.encryption", "aes256"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "esp_ciphers.0.integrity", "sha256"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_connection.main", "esp_ciphers.0.dh_group", "modp2048"),
				),
			},
		},
	})
}

func testAccCheckConnectionExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := s2svpn.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		connection, err := api.GetConnection(&s2s_vpn.GetConnectionRequest{
			ConnectionID: id,
			Region:       region,
		})

		if err != nil {
			return err
		}

		if connection.Status.String() == "error" {
			return fmt.Errorf("connection is in error state")
		}

		return nil
	}
}

func testAccCheckConnectionDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_s2s_vpn_connection" {
				continue
			}

			api, region, id, err := s2svpn.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetConnection(&s2s_vpn.GetConnectionRequest{
				ConnectionID: id,
				Region:       region,
			})

			if err == nil {
				return fmt.Errorf("s2svpn connection (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
