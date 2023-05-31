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

func testSweepVPCPrivateNetwork(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		vpcAPI := vpc.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the private network in (%s)", zone)

		listPNResponse, err := vpcAPI.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing private network in sweeper: %s", err)
		}

		for _, pn := range listPNResponse.PrivateNetworks {
			err := vpcAPI.DeletePrivateNetwork(&vpc.DeletePrivateNetworkRequest{
				Zone:             zone,
				PrivateNetworkID: pn.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting private network in sweeper: %s", err)
			}
		}
		return nil
	})
}

func TestAccScalewayVPCPrivateNetwork_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	privateNetworkName := "private-network-test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc_private_network pn01 {
						name = "%s"
					}
				`, privateNetworkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(
						tt,
						"scaleway_vpc_private_network.pn01",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"name",
						privateNetworkName,
					),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc_private_network pn01 {
						name = "%s"
						tags = ["tag0", "tag1"]
					}
				`, privateNetworkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(
						tt,
						"scaleway_vpc_private_network.pn01",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"tags.0",
						"tag0",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"tags.1",
						"tag1",
					),
				),
			},
		},
	})
}

func TestAccScalewayVPCPrivateNetwork_DefaultName(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `resource scaleway_vpc_private_network main {}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(
						tt,
						"scaleway_vpc_private_network.main",
					),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.main", "name"),
				),
			},
		},
	})
}

func TestAccScalewayVPCPrivateNetwork_Subnets(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `resource scaleway_vpc_private_network test {}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(
						tt,
						"scaleway_vpc_private_network.test",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"subnets.#",
						"2",
					),
				),
			},
			{
				Config: `resource scaleway_vpc_private_network test {
					subnets = ["10.20.0.0/22", "fd21:e31e:86f0:f2c6::/64"]
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(
						tt,
						"scaleway_vpc_private_network.test",
					),
					resource.TestCheckTypeSetElemAttr(
						"scaleway_vpc_private_network.test",
						"subnets.*",
						"10.20.0.0/22",
					),
					resource.TestCheckTypeSetElemAttr(
						"scaleway_vpc_private_network.test",
						"subnets.*",
						"fd21:e31e:86f0:f2c6::/64",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"subnets.#",
						"2",
					),
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

			if err == nil {
				return fmt.Errorf(
					"VPC private network %s still exists",
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
