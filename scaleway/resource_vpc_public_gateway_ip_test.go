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
	resource.AddTestSweepers("scaleway_vpc_public_gateway_ip", &resource.Sweeper{
		Name: "scaleway_vpc_public_gateway_ip",
		F:    testSweepVPCPublicGatewayIP,
	})
}

func testSweepVPCPublicGatewayIP(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		vpcgwAPI := vpcgw.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the public gateways ip in (%s)", zone)

		listIPResponse, err := vpcgwAPI.ListIPs(&vpcgw.ListIPsRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing public gateway ip in sweeper: %w", err)
		}

		for _, ip := range listIPResponse.IPs {
			err := vpcgwAPI.DeleteIP(&vpcgw.DeleteIPRequest{
				Zone: zone,
				IPID: ip.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting public gateway ip in sweeper: %w", err)
			}
		}
		return nil
	})
}

func TestAccScalewayVPCPublicGatewayIP_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	testDNSZone := fmt.Sprintf("%s.%s", testDomainZone, testDomain)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_record" "tf_A" {
						dns_zone = %[1]q
						name     = "tf"
						type     = "A"
						data     = "${scaleway_vpc_public_gateway_ip.main.address}"
						ttl      = 3600
						priority = 1
					}

					resource scaleway_vpc_public_gateway_ip main {
					}
				`, testDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayIPExists(tt, "scaleway_vpc_public_gateway_ip.main"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "reverse"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "address"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "updated_at"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_record" "tf_A" {
						dns_zone = %[1]q
						name     = "tf"
						type     = "A"
						data     = "${scaleway_vpc_public_gateway_ip.main.address}"
						ttl      = 3600
						priority = 1
					}

					resource scaleway_vpc_public_gateway_ip main {
						reverse = %[2]q
						tags = ["tag0", "tag1"]
					}
				`, testDomain, testDNSZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayIPExists(tt, "scaleway_vpc_public_gateway_ip.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip.main", "tags.0", "tag0"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip.main", "tags.1", "tag1"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip.main", "reverse", testDNSZone),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "address"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "updated_at"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_public_gateway_ip main {
						tags = ["tag2", "tag3"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayIPExists(tt, "scaleway_vpc_public_gateway_ip.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip.main", "tags.0", "tag2"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip.main", "tags.1", "tag3"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "reverse"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "address"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "updated_at"),
				),
			},
		},
	})
}

func testAccCheckScalewayVPCPublicGatewayIPExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcgwAPI.GetIP(&vpcgw.GetIPRequest{
			IPID: ID,
			Zone: zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayVPCPublicGatewayIPDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_public_gateway_ip" {
				continue
			}

			vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = vpcgwAPI.GetIP(&vpcgw.GetIPRequest{
				IPID: ID,
				Zone: zone,
			})

			if err == nil {
				return fmt.Errorf(
					"VPC public gateway ip %s still exists",
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
