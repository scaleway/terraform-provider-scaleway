package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_vpc_private_network", &resource.Sweeper{
		Name: "scaleway_vpc_private_network",
		F:    testSweepVPCPrivateNetwork,
	})
}

func testSweepVPCPrivateNetwork(zone string) error {
	scwClient, err := sharedClientForZone(scw.Zone(zone))
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	vpcAPI := vpc.NewAPI(scwClient)

	l.Debugf("sweeper: destroying the private networks in (%s)", zone)
	listPNs, err := vpcAPI.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{}, scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("error listing private networks in (%s) in sweeper: %s", zone, err)
	}

	for _, pn := range listPNs.PrivateNetworks {
		err := vpcAPI.DeletePrivateNetwork(&vpc.DeletePrivateNetworkRequest{
			PrivateNetworkID: pn.ID,
			Zone:             scw.Zone(zone),
		})
		if err != nil {
			return fmt.Errorf("error deleting private network in sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayVPCPrivateNetwork_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network main {
						name = "TestAccScalewayVPCPrivateNetwork_Basic"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(tt, "scaleway_vpc_private_network.main"),
					resource.TestCheckResourceAttr("scaleway_vpc_private_network.main", "name", "TestAccScalewayVPCPrivateNetwork_Basic"),
					testCheckResourceAttrUUID("scaleway_vpc_private_network.main", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayVPCPrivateNetworkExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcAPI, zone, ID, err := vpcAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcAPI.GetPrivateNetwork(&vpc.GetPrivateNetworkRequest{
			PrivateNetworkID: ID,
			Zone:             zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayVPCPrivateNetworkDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_private_network" {
				continue
			}

			vpcAPI, zone, ID, err := vpcAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = vpcAPI.GetPrivateNetwork(&vpc.GetPrivateNetworkRequest{
				PrivateNetworkID: ID,
				Zone:             zone,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("private network (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
