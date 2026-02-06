package s2svpn_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceVPNGateway_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckVPNGatewayDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-test-vpc-vpn-gateway-ds"
					}

					resource "scaleway_vpc_private_network" "main" {
						vpc_id = scaleway_vpc.main.id
						ipv4_subnet {
							subnet = "10.0.0.0/24"
						}
					}

					resource "scaleway_s2s_vpn_gateway" "main" {
						name              = "tf-test-vpn-gateway-ds"
						gateway_type      = "VGW-S"
						private_network_id = scaleway_vpc_private_network.main.id
						region            = "fr-par"
						zone              = "fr-par-1"
					}
				`,
			},
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-test-vpc-vpn-gateway-ds"
					}

					resource "scaleway_vpc_private_network" "main" {
						vpc_id = scaleway_vpc.main.id
						ipv4_subnet {
							subnet = "10.0.0.0/24"
						}
					}

					resource "scaleway_s2s_vpn_gateway" "main" {
						name              = "tf-test-vpn-gateway-ds"
						gateway_type      = "VGW-S"
						private_network_id = scaleway_vpc_private_network.main.id
						region            = "fr-par"
						zone              = "fr-par-1"
					}

					data "scaleway_s2s_vpn_gateway" "by_name" {
						name = scaleway_s2s_vpn_gateway.main.name
					}

					data "scaleway_s2s_vpn_gateway" "by_id" {
						vpn_gateway_id = scaleway_s2s_vpn_gateway.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNGatewayExists(tt, "scaleway_s2s_vpn_gateway.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_gateway.by_name", "name",
						"scaleway_s2s_vpn_gateway.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_gateway.by_name", "gateway_type",
						"scaleway_s2s_vpn_gateway.main", "gateway_type"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_gateway.by_id", "vpn_gateway_id",
						"scaleway_s2s_vpn_gateway.main", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_gateway.by_id", "name",
						"scaleway_s2s_vpn_gateway.main", "name"),
				),
			},
		},
	})
}
