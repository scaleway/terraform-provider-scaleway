package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_gateway_network", &resource.Sweeper{
		Name: "scaleway_gateway_network",
		F:    testSweepVPCGatewayNetwork,
	})
}

func testSweepVPCGatewayNetwork(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		vpcgwAPI := vpcgw.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the gateway network in (%s)", zone)

		listPNResponse, err := vpcgwAPI.ListGatewayNetworks(&vpcgw.ListGatewayNetworksRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing gateway network in sweeper: %s", err)
		}

		for _, gn := range listPNResponse.GatewayNetworks {
			err := vpcgwAPI.DeleteGatewayNetwork(&vpcgw.DeleteGatewayNetworkRequest{
				GatewayNetworkID: gn.GatewayID,
				Zone:             zone,
				// Cleanup the dhcp resource related. DON'T CALL THE SWEEPER DHCP
				CleanupDHCP: true,
			})
			if err != nil {
				return fmt.Errorf("error deleting gateway network in sweeper: %s", err)
			}
		}
		return nil
	})
}

func TestAccScalewayVPCGatewayNetwork_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCGatewayNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "pn_test_network"
					}

					resource scaleway_vpc_public_gateway_ip gw01 {
					}

					resource scaleway_vpc_public_gateway_dhcp dhcp01 {
						subnet = "192.168.1.0/24"
					}
				`,
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "pn_test_network"
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
						private_network_id = scaleway_vpc_private_network.pn01.id
						dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
						cleanup_dhcp = true
						enable_masquerade = true
						depends_on = [scaleway_vpc_public_gateway_ip.gw01, scaleway_vpc_private_network.pn01]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCGatewayNetworkExists(tt, "scaleway_vpc_gateway_network.main"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "gateway_id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "private_network_id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "dhcp_id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "zone"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "enable_dhcp", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "cleanup_dhcp", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "enable_masquerade", "true"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "pn_test_network"
					}

					resource scaleway_vpc_public_gateway_ip gw01 {
					}

					resource scaleway_vpc_public_gateway pg01 {
						name = "foobar"
						type = "VPC-GW-S"
						ip_id = scaleway_vpc_public_gateway_ip.gw01.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pn01", "name"),
				),
			},
		},
	})
}

func TestAccScalewayVPCGatewayNetwork_WithoutDHCP(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCGatewayNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "pn_test_network"
					}

					resource scaleway_vpc_public_gateway pg01 {
						name = "foobar"
						type = "VPC-GW-S"
					}

					resource scaleway_vpc_gateway_network main {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.pn01.id
						enable_dhcp = false
						enable_masquerade = true
						static_address = "192.168.1.42/24"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCGatewayNetworkExists(tt, "scaleway_vpc_gateway_network.main"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "gateway_id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "private_network_id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "zone"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "static_address", "192.168.1.42/24"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "enable_dhcp", "false"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "cleanup_dhcp", "false"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "enable_masquerade", "true"),
				),
			},
		},
	})
}

func testAccCheckScalewayVPCGatewayNetworkExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcgwNetworkAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcgwNetworkAPI.GetGatewayNetwork(&vpcgw.GetGatewayNetworkRequest{
			GatewayNetworkID: ID,
			Zone:             zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayVPCGatewayNetworkDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_gateway_network" {
				continue
			}

			vpcgwNetworkAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = vpcgwNetworkAPI.GetGatewayNetwork(&vpcgw.GetGatewayNetworkRequest{
				GatewayNetworkID: ID,
				Zone:             zone,
			})

			if err == nil {
				return fmt.Errorf(
					"VPC gateway network %s still exists",
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
