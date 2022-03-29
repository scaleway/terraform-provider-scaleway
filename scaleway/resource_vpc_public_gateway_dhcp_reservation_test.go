package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
)

func TestAccScalewayVPCPublicGatewayDHCPEntry_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayDHCPEntryDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network internal {
						name = "pn_test_network"
					}

					resource "scaleway_instance_server" "main" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						zone = "fr-par-1"

						private_network {
							pn_id = scaleway_vpc_private_network.internal.id
						}
					}

					resource scaleway_vpc_public_gateway_ip gw01 {
					}

					resource scaleway_vpc_public_gateway_dhcp dhcp01 {
						subnet = "192.168.1.0/24"
					}

					resource scaleway_vpc_public_gateway pg01 {
						name = "foobar"
						type = "VPC-GW-S"
						ip_id = scaleway_vpc_public_gateway_ip.gw01.id
					}

					resource scaleway_vpc_gateway_network main {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.internal.id
						dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
						cleanup_dhcp = true
						enable_masquerade = true
						depends_on = [scaleway_vpc_public_gateway_ip.gw01, scaleway_vpc_private_network.internal]
					}

					resource scaleway_vpc_public_gateway_dhcp_reservation main {
						gateway_network_id = scaleway_vpc_gateway_network.id
						mac_address = scaleway_instance_server.private_network.0.mac_address
						ip_address = "192.168.1.1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayDHCPReservationExists(tt, "scaleway_vpc_public_gateway_dhcp_reservation.main"),
					resource.TestCheckResourceAttrPair("scaleway_vpc_public_gateway_dhcp_reservation",
						"mac_address", "scaleway_instance_server", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrPair("scaleway_vpc_public_gateway_dhcp_reservation", "gateway_network_id",
						"scaleway_vpc_gateway_network", "id"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp_reservation", "ip_address", "192.168.1.1"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp_reservation", "hostname"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp_reservation", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp_reservation", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp_reservation", "type"),
				),
			},
		},
	})
}

func testAccCheckScalewayVPCPublicGatewayDHCPReservationExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcgwAPI.GetDHCPEntry(&vpcgw.GetDHCPEntryRequest{
			DHCPEntryID: ID,
			Zone:        zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayVPCPublicGatewayDHCPEntryDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_public_gateway_dhcp_reservation" {
				continue
			}

			vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = vpcgwAPI.GetDHCPEntry(&vpcgw.GetDHCPEntryRequest{
				DHCPEntryID: ID,
				Zone:        zone,
			})

			if err == nil {
				return fmt.Errorf(
					"VPC public gateway DHCP Entry config %s still exists",
					rs.Primary.ID,
				)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
