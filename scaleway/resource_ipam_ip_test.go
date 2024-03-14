package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func init() {
	resource.AddTestSweepers("scaleway_ipam_ip", &resource.Sweeper{
		Name: "scaleway_ipam_ip",
		F:    testSweepIPAMIP,
	})
}

func testSweepIPAMIP(_ string) error {
	return sweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		ipamAPI := ipam.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting the IPs in (%s)", region)

		listIPs, err := ipamAPI.ListIPs(&ipam.ListIPsRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing ips in (%s) in sweeper: %s", region, err)
		}

		for _, v := range listIPs.IPs {
			err := ipamAPI.ReleaseIP(&ipam.ReleaseIPRequest{
				IPID:   v.ID,
				Region: region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error releasing IP in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayIPAMIP_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIPAMIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "my vpc"
					}

					resource scaleway_vpc_private_network pn01 {
						vpc_id = scaleway_vpc.vpc01.id
						ipv4_subnet {
							subnet = "172.16.32.0/22"
						}
					}

					resource scaleway_ipam_ip ip01 {
					 source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPAMIPExists(tt, "scaleway_ipam_ip.ip01"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip.ip01", "is_ipv6", "false"),
					acctest.CheckResourceRawIDMatches("scaleway_ipam_ip.ip01", "source.0.private_network_id", "scaleway_vpc_private_network.pn01", "id"),
					resource.TestCheckResourceAttrSet("scaleway_ipam_ip.ip01", "source.0.subnet_id"),
					resource.TestCheckResourceAttrSet("scaleway_ipam_ip.ip01", "address"),
					resource.TestCheckResourceAttrSet("scaleway_ipam_ip.ip01", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_ipam_ip.ip01", "updated_at"),
				),
			},
		},
	})
}

func TestAccScalewayIPAMIP_WithStandaloneAddress(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIPAMIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "my vpc"
					}

					resource scaleway_vpc_private_network pn01 {
						vpc_id = scaleway_vpc.vpc01.id
						ipv4_subnet {
							subnet = "172.16.32.0/22"
						}
					}

					resource scaleway_ipam_ip ip01 {
					 address = "172.16.32.7"
					 source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPAMIPExists(tt, "scaleway_ipam_ip.ip01"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip.ip01", "address", "172.16.32.7/22"),
				),
			},
		},
	})
}

func TestAccScalewayIPAMIP_WithTags(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIPAMIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "my vpc"
					}

					resource scaleway_vpc_private_network pn01 {
						vpc_id = scaleway_vpc.vpc01.id
						ipv4_subnet {
							subnet = "172.16.32.0/22"
						}
					}

					resource scaleway_ipam_ip ip01 {
					 source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					  tags = [ "terraform-test", "ipam" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPAMIPExists(tt, "scaleway_ipam_ip.ip01"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip.ip01", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip.ip01", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip.ip01", "tags.1", "ipam"),
				),
			},
			{
				Config: `
					resource scaleway_vpc vpc01 {
						name = "my vpc"
					}

					resource scaleway_vpc_private_network pn01 {
						vpc_id = scaleway_vpc.vpc01.id
						ipv4_subnet {
							subnet = "172.16.32.0/22"
						}
					}

					resource scaleway_ipam_ip ip01 {
					 source {
						private_network_id = scaleway_vpc_private_network.pn01.id
					  }
					  tags = [ "terraform-test", "ipam", "updated" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPAMIPExists(tt, "scaleway_ipam_ip.ip01"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip.ip01", "tags.#", "3"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip.ip01", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip.ip01", "tags.1", "ipam"),
					resource.TestCheckResourceAttr("scaleway_ipam_ip.ip01", "tags.2", "updated"),
				),
			},
		},
	})
}

func testAccCheckScalewayIPAMIPExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		ipamAPI, region, ID, err := scaleway.IpamAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = ipamAPI.GetIP(&ipam.GetIPRequest{
			IPID:   ID,
			Region: region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayIPAMIPDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_ipam_ip" {
				continue
			}

			ipamAPI, region, ID, err := scaleway.IpamAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = ipamAPI.GetIP(&ipam.GetIPRequest{
				IPID:   ID,
				Region: region,
			})

			if err == nil {
				return fmt.Errorf("IP (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
