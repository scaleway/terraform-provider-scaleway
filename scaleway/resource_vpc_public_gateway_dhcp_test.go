package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
)

func TestAccScalewayVPCPublicGatewayDHCP_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayDHCPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_public_gateway_dhcp main {
						subnet = "192.168.1.0/24"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayDHCPExists(tt, "scaleway_vpc_public_gateway_dhcp.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "subnet", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "enable_dynamic", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "valid_lifetime", "3600"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "renew_timer", "3000"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "rebind_timer", "3060"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_default_route", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_dns_server", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_server_override.#", "0"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_search.#", "0"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "dns_local_name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_low"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_high"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "zone"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "organization_id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_public_gateway_dhcp main {
						subnet = "192.168.1.0/24"
						valid_lifetime = 3000
						renew_timer = 2000
						rebind_timer = 2060
						push_default_route = false
						push_dns_server = false
						enable_dynamic = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayDHCPExists(tt, "scaleway_vpc_public_gateway_dhcp.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "subnet", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_default_route", "false"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_dns_server", "false"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "enable_dynamic", "false"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "valid_lifetime", "3000"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "renew_timer", "2000"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "rebind_timer", "2060"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_server_override.#", "0"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_search.#", "0"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "dns_local_name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_low"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_high"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "zone"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "organization_id"),
				),
			},
		},
	})
}

func TestAccScalewayVPCPublicGatewayDHCP_Basic2(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayDHCPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_public_gateway_dhcp" main {
					  subnet = "192.168.1.0/24"
					  push_default_route = true
					  push_dns_server = true
					  enable_dynamic = true
					  dns_servers_override = ["192.168.1.2"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayDHCPExists(tt, "scaleway_vpc_public_gateway_dhcp.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "subnet", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "enable_dynamic", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "valid_lifetime", "3600"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "renew_timer", "3000"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "rebind_timer", "3060"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_default_route", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_dns_server", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_servers_override.#", "1"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_servers_override.0", "192.168.1.2"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "dns_local_name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_low"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_high"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "zone"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "organization_id"),
				),
			},
			{
				Config: `
					resource "scaleway_vpc_public_gateway_dhcp" main {
					  subnet = "192.168.1.0/24"
					  push_default_route = true
					  push_dns_server = true
					  enable_dynamic = true
					  dns_servers_override = ["192.168.1.3"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayDHCPExists(tt, "scaleway_vpc_public_gateway_dhcp.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "subnet", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "enable_dynamic", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "valid_lifetime", "3600"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "renew_timer", "3000"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "rebind_timer", "3060"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_default_route", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_dns_server", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_servers_override.#", "1"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_servers_override.0", "192.168.1.3"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "dns_local_name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_low"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_high"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "zone"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "organization_id"),
				),
			},
		},
	})
}

func TestAccScalewayVPCPublicGatewayDHCP_WithReservation(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayDHCPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_private_network" main {
					  name = "subnet_all"
					}

					resource "scaleway_vpc_public_gateway_ip" "main" {}

					resource "scaleway_vpc_public_gateway_dhcp" "main" {
					  subnet = "192.168.1.0/24"
					  push_default_route = true
					  push_dns_server = true
					  enable_dynamic = true
					  dns_servers_override = ["192.168.1.2"]
					}

					resource "scaleway_vpc_public_gateway" "main" {
					  name = "public gateway"
					  type = "VPC-GW-S"
					  ip_id = scaleway_vpc_public_gateway_ip.main.id
					}

					resource "scaleway_vpc_gateway_network" "main" {
					  gateway_id = scaleway_vpc_public_gateway.main.id
					  private_network_id = scaleway_vpc_private_network.main.id
					  dhcp_id = scaleway_vpc_public_gateway_dhcp.main.id
					  cleanup_dhcp = true
					  enable_masquerade = true
					  enable_dhcp = true
					  depends_on = [scaleway_vpc_public_gateway_ip.main, scaleway_vpc_private_network.main]
					}

					resource "scaleway_instance_server" "main" {
					  count = 1
					  name = "front-node${count.index+1}"
					  type = "DEV1-S"
					  image = "debian_bullseye"
					  private_network {
						pn_id = scaleway_vpc_private_network.main.id
					  }
					}
					
					resource scaleway_vpc_public_gateway_dhcp_reservation main {
					  count = 1
					  gateway_network_id = scaleway_vpc_gateway_network.main.id
					  mac_address = scaleway_instance_server.main[count.index].private_network[0].mac_address
					  ip_address = "192.168.1.${count.index+101}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayDHCPExists(tt, "scaleway_vpc_public_gateway_dhcp.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "subnet", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "enable_dynamic", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "valid_lifetime", "3600"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "renew_timer", "3000"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "rebind_timer", "3060"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_default_route", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "push_dns_server", "true"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_servers_override.#", "1"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "dns_servers_override.0", "192.168.1.2"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "dns_local_name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_low"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "pool_high"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "zone"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.main", "organization_id"),
				),
			},
		},
	})
}

func TestAccScalewayVPCPublicGatewayDHCP_WithPatRule(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayDHCPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					### Scaleway Private Network
					resource "scaleway_vpc_private_network" "main" {
					  name = "Monitoring"
					}
					
					### IP for Public Gateway
					resource "scaleway_vpc_public_gateway_ip" "main" {
					}
					
					### DHCP Space of VPC
					resource "scaleway_vpc_public_gateway_dhcp" "main" {
					  subnet = "10.0.0.0/24"
					}
					
					### The Public Gateway with the Attached IP
					resource "scaleway_vpc_public_gateway" "main" {
					  name  = "Monitoring"
					  type  = "VPC-GW-S"
					  ip_id = scaleway_vpc_public_gateway_ip.main.id
					}
					
					### VPC Gateway Network
					resource "scaleway_vpc_gateway_network" "main" {
					  gateway_id         = scaleway_vpc_public_gateway.main.id
					  private_network_id = scaleway_vpc_private_network.main.id
					  dhcp_id            = scaleway_vpc_public_gateway_dhcp.main.id
					  cleanup_dhcp       = true
					  enable_masquerade  = true
					  depends_on = [
						scaleway_vpc_public_gateway_ip.main,
						scaleway_vpc_private_network.main
					  ]
					}
					
					### Elastic Stack Instance
					resource "scaleway_instance_server" "main" {
					  name        = "Elastic Stack"
					  type        = "DEV1-S"
					  image       = "debian_bullseye"
					  enable_ipv6 = false
					
					  private_network {
						pn_id = scaleway_vpc_private_network.main.id
					  }
					}
					
					### DHCP Reservation for Elastic Stack Instance
					resource "scaleway_vpc_public_gateway_dhcp_reservation" "main" {
					  gateway_network_id = scaleway_vpc_gateway_network.main.id
					  mac_address        = scaleway_instance_server.main.private_network.0.mac_address
					  ip_address         = "10.0.0.3"
					}
					
					### Elastic Stack SSH Port
					resource "scaleway_vpc_public_gateway_pat_rule" "main" {
					  gateway_id   = scaleway_vpc_public_gateway.main.id
					  private_ip   = scaleway_vpc_public_gateway_dhcp_reservation.main.ip_address
					  private_port = 22
					  public_port  = 2023
					  protocol     = "tcp"
					  depends_on = [
						scaleway_vpc_gateway_network.main,
						scaleway_vpc_private_network.main
					  ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayDHCPExists(tt, "scaleway_vpc_public_gateway_dhcp.main"),
					testAccCheckScalewayVPCPublicGatewayPATRuleExists(tt, "scaleway_vpc_public_gateway_pat_rule.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_dhcp.main", "subnet", "10.0.0.0/24"),
					resource.TestCheckResourceAttrPair(
						"scaleway_vpc_public_gateway_pat_rule.main", "private_ip",
						"scaleway_vpc_public_gateway_dhcp_reservation.main", "ip_address"),
				),
			},
		},
	})
}

func testAccCheckScalewayVPCPublicGatewayDHCPExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcgwAPI.GetDHCP(&vpcgw.GetDHCPRequest{
			DHCPID: ID,
			Zone:   zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayVPCPublicGatewayDHCPDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_public_gateway_dhcp" {
				continue
			}

			vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = vpcgwAPI.GetDHCP(&vpcgw.GetDHCPRequest{
				DHCPID: ID,
				Zone:   zone,
			})

			if err == nil {
				return fmt.Errorf(
					"VPC public gateway DHCP config %s still exists",
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
