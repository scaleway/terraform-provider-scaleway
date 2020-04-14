package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_lb_ip_beta", &resource.Sweeper{
		Name: "scaleway_lb_ip_beta",
		F:    testSweepLBIP,
	})
}

func testSweepLBIP(region string) error {
	scwClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	lbAPI := lb.NewAPI(scwClient)

	l.Debugf("sweeper: destroying the lb ips in (%s)", region)
	listIPs, err := lbAPI.ListIPs(&lb.ListIPsRequest{}, scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("error listing lb ips in (%s) in sweeper: %s", region, err)
	}

	for _, ip := range listIPs.IPs {
		if ip.LbID == nil {
			err := lbAPI.ReleaseIP(&lb.ReleaseIPRequest{
				IPID: ip.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting lb ip in sweeper: %s", err)
			}
		}
	}

	return nil
}

func TestAccScalewayLbIPBeta(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayLbIPBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPBetaExists("scaleway_lb_ip_beta.ip01"),
					testCheckResourceAttrIPv4("scaleway_lb_ip_beta.ip01", "ip_address"),
					resource.TestCheckResourceAttrSet("scaleway_lb_ip_beta.ip01", "reverse"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {
						reverse = "myreverse.com"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPBetaExists("scaleway_lb_ip_beta.ip01"),
					testCheckResourceAttrIPv4("scaleway_lb_ip_beta.ip01", "ip_address"),
					resource.TestCheckResourceAttr("scaleway_lb_ip_beta.ip01", "reverse", "myreverse.com"),
				),
			},
		},
	})
}

func testAccCheckScalewayLbIPBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, region, ID, err := lbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetIP(&lb.GetIPRequest{
			IPID:   ID,
			Region: region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayLbIPBetaDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_lb_ip_beta" {
			continue
		}

		lbAPI, region, ID, err := lbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetIP(&lb.GetIPRequest{
			Region: region,
			IPID:   ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("IP (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}
