package vpcgw_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpcgwSDK "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw"
)

func TestAccVPCPublicGatewayPATRule_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckVPCPublicGatewayPATRuleDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					### Scaleway Private Network
					resource "scaleway_vpc_private_network" "main" {
					  name = "My Private Network"
					}
					
					### IP for Public Gateway
					resource "scaleway_vpc_public_gateway_ip" "main" {
					}
	
					### The Public Gateway with the Attached IP
					resource "scaleway_vpc_public_gateway" "main" {
					  name  = "The Public Gateway"
					  type  = "VPC-GW-S"
					  ip_id = scaleway_vpc_public_gateway_ip.main.id
					}
					
					### VPC Gateway Network
					resource "scaleway_vpc_gateway_network" "main" {
					  gateway_id         = scaleway_vpc_public_gateway.main.id
					  private_network_id = scaleway_vpc_private_network.main.id
					  enable_masquerade  = true
					  ipam_config {
						push_default_route = false
					  }
					}
					
					### Scaleway Instance
					resource "scaleway_instance_server" "main" {
					  name        = "Scaleway Instance"
					  type        = "DEV1-S"
					  image       = "debian_bullseye"
					
					  private_network {
						pn_id = scaleway_vpc_private_network.main.id
					  }
					}

					data "scaleway_ipam_ip" "main" {
					  mac_address = scaleway_instance_server.main.private_network.0.mac_address
					  type        = "ipv4"
					}
					
					### PAT RULE for SSH Port
					resource "scaleway_vpc_public_gateway_pat_rule" "main" {
					  gateway_id   = scaleway_vpc_public_gateway.main.id
					  private_ip   = data.scaleway_ipam_ip.main.address
					  private_port = 22
					  public_port  = 2023
					  protocol     = "tcp"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPublicGatewayPATRuleExists(tt, "scaleway_vpc_public_gateway_pat_rule.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_pat_rule.main", "protocol", "tcp"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_pat_rule.main", "public_port", "2023"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_pat_rule.main", "private_port", "22"),
					resource.TestCheckResourceAttrPair(
						"scaleway_vpc_public_gateway_pat_rule.main", "private_ip",
						"data.scaleway_ipam_ip.main", "address"),
					resource.TestCheckResourceAttrPair(
						"scaleway_vpc_public_gateway_pat_rule.main", "gateway_id",
						"scaleway_vpc_public_gateway.main", "id"),
				),
			},
		},
	})
}

func testAccCheckVPCPublicGatewayPATRuleExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, ID, err := vpcgw.NewAPIWithZoneAndIDv2(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetPatRule(&vpcgwSDK.GetPatRuleRequest{
			PatRuleID: ID,
			Zone:      zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckVPCPublicGatewayPATRuleDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_public_gateway_pat_rules" {
				continue
			}

			api, zone, ID, err := vpcgw.NewAPIWithZoneAndIDv2(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetPatRule(&vpcgwSDK.GetPatRuleRequest{
				PatRuleID: ID,
				Zone:      zone,
			})

			if err == nil {
				return fmt.Errorf(
					"VPC public gateway pat rule %s still exists",
					rs.Primary.ID,
				)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
