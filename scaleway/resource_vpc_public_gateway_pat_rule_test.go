package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1beta1"
)

func init() {
	resource.AddTestSweepers("scaleway_vpc_public_gateway_pat_rule", &resource.Sweeper{
		Name: "scaleway_vpc_public_gateway_pat_rule",
		F:    testSweepVPCPublicGateway,
	})
}

func TestAccScalewayVPCPublicGatewayPATRule_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayPATRuleDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_public_gateway pg01 {
						type = "VPC-GW-S"
					}

					resource scaleway_vpc_public_gateway_dhcp dhcp01 {
						subnet = "192.168.1.0/24"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "pn_test_network"
					}

					resource scaleway_vpc_gateway_network gn01 {
					    gateway_id = scaleway_vpc_public_gateway.pg01.id
					    private_network_id = scaleway_vpc_private_network.pn01.id
					    dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
					}

					resource scaleway_vpc_public_gateway_pat_rule main {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_ip = scaleway_vpc_public_gateway_dhcp.dhcp01.address
						private_port = 42
						public_port = 42
						protocol = "both"
						depends_on = [scaleway_vpc_gateway_network.gn01]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayPATRuleExists(
						tt,
						"scaleway_vpc_public_gateway_pat_rule.main",
					),
				),
			},
		},
	})
}

func testAccCheckScalewayVPCPublicGatewayPATRuleExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcgwAPI.GetPATRule(&vpcgw.GetPATRuleRequest{
			PatRuleID: ID,
			Zone:      zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayVPCPublicGatewayPATRuleDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_public_gateway_pat_rules" {
				continue
			}

			vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = vpcgwAPI.GetPATRule(&vpcgw.GetPATRuleRequest{
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
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
