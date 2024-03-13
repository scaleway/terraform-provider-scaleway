package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
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
		logging.L.Debugf("sweeper: destroying the public gateways ip in (%s)", zone)

		listIPResponse, err := vpcgwAPI.ListIPs(&vpcgw.ListIPsRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing public gateway ip in sweeper: %s", err)
		}

		for _, ip := range listIPResponse.IPs {
			err := vpcgwAPI.DeleteIP(&vpcgw.DeleteIPRequest{
				Zone: zone,
				IPID: ip.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting public gateway ip in sweeper: %s", err)
			}
		}
		return nil
	})
}

func TestAccScalewayVPCPublicGatewayIP_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_public_gateway_ip main {}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayIPExists(tt, "scaleway_vpc_public_gateway_ip.main"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "address"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.main", "updated_at"),
				),
			},
		},
	})
}

func TestAccScalewayVPCPublicGatewayIP_WithZone(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_public_gateway_ip main {}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayIPExists(tt, "scaleway_vpc_public_gateway_ip.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip.main", "zone", "fr-par-1"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_public_gateway_ip main {
						zone = "nl-ams-1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayIPExists(tt, "scaleway_vpc_public_gateway_ip.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip.main", "zone", "nl-ams-1"),
				),
			},
		},
	})
}

func TestAccScalewayVPCPublicGatewayIP_WithTags(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPublicGatewayIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_public_gateway_ip main {}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayIPExists(tt, "scaleway_vpc_public_gateway_ip.main"),
					resource.TestCheckNoResourceAttr("scaleway_vpc_public_gateway_ip.main", "tags"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_public_gateway_ip main {
						tags = ["foo", "bar"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPublicGatewayIPExists(tt, "scaleway_vpc_public_gateway_ip.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip.main", "tags.0", "foo"),
					resource.TestCheckResourceAttr("scaleway_vpc_public_gateway_ip.main", "tags.1", "bar"),
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
			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
