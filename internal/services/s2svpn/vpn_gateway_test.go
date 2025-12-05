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

func TestAccVPNGateway_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckVPNGatewayDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-test-vpc-vpn-gateway"
					}

					resource "scaleway_vpc_private_network" "main" {
						vpc_id = scaleway_vpc.main.id
						ipv4_subnet {
							subnet = "10.0.0.0/24"
						}
					}

					resource "scaleway_s2s_vpn_gateway" "main" {
						name              = "tf-test-vpn-gateway"
						gateway_type      = "VGW-S"
						private_network_id = scaleway_vpc_private_network.main.id
						region            = "fr-par"
						zone              = "fr-par-1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNGatewayExists(tt, "scaleway_s2s_vpn_gateway.main"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_gateway.main", "name", "tf-test-vpn-gateway"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_gateway.main", "gateway_type", "VGW-S"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_gateway.main", "region", "fr-par"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_gateway.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_gateway.main", "asn"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_gateway.main", "status"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_gateway.main", "public_config.#"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_gateway.main", "ipam_private_ipv4_id"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_gateway.main", "ipam_private_ipv6_id"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_gateway.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_gateway.main", "updated_at"),
				),
			},
		},
	})
}

func testAccCheckVPNGatewayExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := s2svpn.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetVpnGateway(&s2s_vpn.GetVpnGatewayRequest{
			GatewayID: id,
			Region:    region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckVPNGatewayDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_s2s_vpn_gateway" {
				continue
			}

			api, region, id, err := s2svpn.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetVpnGateway(&s2s_vpn.GetVpnGatewayRequest{
				GatewayID: id,
				Region:    region,
			})

			if err == nil {
				return fmt.Errorf("VPN gateway (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
