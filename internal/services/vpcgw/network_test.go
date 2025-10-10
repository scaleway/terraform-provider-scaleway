package vpcgw_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	vpcgwSDK "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	ipamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw"
	vpcgwchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw/testfuncs"
)

func TestAccVPCGatewayNetwork_WithIPAMConfig(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			vpcgwchecks.IsGatewayNetworkDestroyed(tt),
			ipamchecks.CheckIPDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "my vpc"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "pn_test_network"
						ipv4_subnet {
							subnet = "172.16.64.0/22"
						}
						vpc_id = scaleway_vpc.vpc01.id
					}

					resource scaleway_vpc_public_gateway pg01 {
						name = "foobar"
						type = "VPC-GW-S"
					}

					resource scaleway_vpc_gateway_network main {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.pn01.id
						enable_masquerade = true
						ipam_config {
							push_default_route = true
						}					
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCGatewayNetworkExists(tt, "scaleway_vpc_gateway_network.main"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "gateway_id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "private_network_id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "updated_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "status"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "zone"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "ipam_config.0.push_default_route", "true"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "ipam_config.0.ipam_ip_id"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "enable_masquerade", "true"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "private_ip.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "private_ip.0.address"),
				),
			},
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "my vpc"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "pn_test_network"
						ipv4_subnet {
							subnet = "172.16.64.0/22"
						}
						vpc_id = scaleway_vpc.vpc01.id
					}

					resource scaleway_vpc_public_gateway pg01 {
						name = "foobar"
						type = "VPC-GW-S"
					}

					resource "scaleway_ipam_ip" "ip01" {
					  address = "172.16.64.7"
					  source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}

					resource scaleway_vpc_gateway_network main {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.pn01.id
						enable_masquerade = true
						ipam_config {
							push_default_route = true
							ipam_ip_id = scaleway_ipam_ip.ip01.id
						}					
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCGatewayNetworkExists(tt, "scaleway_vpc_gateway_network.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.main", "ipam_config.0.push_default_route", "true"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.main", "ipam_config.0.ipam_ip_id"),
					acctest.CheckResourceRawIDMatches("scaleway_vpc_gateway_network.main", "ipam_config.0.ipam_ip_id", "scaleway_ipam_ip.ip01", "id"),
				),
			},
		},
	})
}

func testAccCheckVPCGatewayNetworkExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, ID, err := vpcgw.NewAPIWithZoneAndIDv2(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetGatewayNetwork(&vpcgwSDK.GetGatewayNetworkRequest{
			GatewayNetworkID: ID,
			Zone:             zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
