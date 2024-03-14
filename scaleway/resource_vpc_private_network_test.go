package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func init() {
	resource.AddTestSweepers("scaleway_vpc_private_network", &resource.Sweeper{
		Name:         "scaleway_vpc_private_network",
		F:            testSweepVPCPrivateNetwork,
		Dependencies: []string{"scaleway_ipam_ip"},
	})
}

func testSweepVPCPrivateNetwork(_ string) error {
	err := sweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		vpcAPI := v2.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the private network in (%s)", region)

		listPNResponse, err := vpcAPI.ListPrivateNetworks(&v2.ListPrivateNetworksRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing private network in sweeper: %s", err)
		}

		for _, pn := range listPNResponse.PrivateNetworks {
			err := vpcAPI.DeletePrivateNetwork(&v2.DeletePrivateNetworkRequest{
				Region:           region,
				PrivateNetworkID: pn.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting private network in sweeper: %s", err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func TestAccScalewayVPCPrivateNetwork_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	privateNetworkName := "private-network-test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
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
					resource.TestCheckResourceAttrSet(
						"scaleway_vpc_private_network.pn01",
						"vpc_id"),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"name",
						privateNetworkName,
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"region",
						"fr-par",
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
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
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
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "my vpc"
					}

					resource scaleway_vpc_private_network test {
						vpc_id = scaleway_vpc.vpc01.id
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(
						tt,
						"scaleway_vpc_private_network.test",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv4_subnet.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv6_subnets.#",
						"1",
					),
				),
			},
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "my vpc"
					}
					
					resource scaleway_vpc_private_network test {
						ipv4_subnet {
							subnet = "172.16.32.0/22"
						}
						ipv6_subnets {
							subnet = "fd46:78ab:30b8:177c::/64"
						}
						vpc_id = scaleway_vpc.vpc01.id
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(
						tt,
						"scaleway_vpc_private_network.test",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv4_subnet.0.subnet",
						"172.16.32.0/22",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv4_subnet.0.address",
						"172.16.32.0",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv4_subnet.0.subnet_mask",
						"255.255.252.0",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv4_subnet.0.prefix_length",
						"22",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv6_subnets.0.subnet",
						"fd46:78ab:30b8:177c::/64",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv6_subnets.0.address",
						"fd46:78ab:30b8:177c::",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv6_subnets.0.prefix_length",
						"64",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv4_subnet.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv6_subnets.#",
						"1",
					),
				),
			},
		},
	})
}

func TestAccScalewayVPCPrivateNetwork_OneSubnet(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "my vpc"
					}

					resource scaleway_vpc_private_network test {
						ipv4_subnet {
							subnet = "172.16.64.0/22"
						}
						vpc_id = scaleway_vpc.vpc01.id
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(
						tt,
						"scaleway_vpc_private_network.test",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv4_subnet.0.subnet",
						"172.16.64.0/22",
					),
					resource.TestCheckResourceAttrSet(
						"scaleway_vpc_private_network.test",
						"ipv6_subnets.0.subnet",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv4_subnet.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.test",
						"ipv6_subnets.#",
						"1",
					),
				),
			},
		},
	})
}

func TestAccScalewayVPCPrivateNetwork_WithTwoIPV6Subnets(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "test-vpc"
						tags = [ "terraform-test", "vpc", "update" ]
					}
					
					resource scaleway_vpc_private_network pn01 {
						name = "pn1"
						tags = ["tag0", "tag1"]
						vpc_id = scaleway_vpc.vpc01.id
						ipv4_subnet {
						  subnet = "192.168.0.0/24"
						}
						ipv6_subnets {
						  subnet = "fd46:78ab:30b8:177c::/64"
						}
						ipv6_subnets {
						  subnet = "fd46:78ab:30b8:c7df::/64"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVPCPrivateNetworkExists(
						tt,
						"scaleway_vpc_private_network.pn01",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"ipv4_subnet.0.subnet",
						"192.168.0.0/24",
					),
					resource.TestCheckResourceAttrSet(
						"scaleway_vpc_private_network.pn01",
						"ipv6_subnets.0.subnet",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"scaleway_vpc_private_network.pn01", "ipv6_subnets.*", map[string]string{
							"subnet": "fd46:78ab:30b8:177c::/64",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"scaleway_vpc_private_network.pn01", "ipv6_subnets.*", map[string]string{
							"subnet": "fd46:78ab:30b8:c7df::/64",
						}),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"ipv4_subnet.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"ipv6_subnets.#",
						"2",
					),
				),
			},
		},
	})
}

func testAccCheckScalewayVPCPrivateNetworkExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcAPI, region, ID, err := scaleway.VpcAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcAPI.GetPrivateNetwork(&v2.GetPrivateNetworkRequest{
			PrivateNetworkID: ID,
			Region:           region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayVPCPrivateNetworkDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_vpc_private_network" {
				continue
			}

			vpcAPI, region, ID, err := scaleway.VpcAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}
			_, err = vpcAPI.GetPrivateNetwork(&v2.GetPrivateNetworkRequest{
				PrivateNetworkID: ID,
				Region:           region,
			})

			if err == nil {
				return fmt.Errorf(
					"VPC private network %s still exists",
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
