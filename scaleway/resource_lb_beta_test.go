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
	resource.AddTestSweepers("scaleway_lb_beta", &resource.Sweeper{
		Name: "scaleway_lb_beta",
		F:    testSweepLB,
	})
}

func testSweepLB(region string) error {
	scwClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	lbAPI := lb.NewAPI(scwClient)

	l.Debugf("sweeper: destroying the lbs in (%s)", region)
	listLBs, err := lbAPI.ListLbs(&lb.ListLbsRequest{}, scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("error listing lbs in (%s) in sweeper: %s", region, err)
	}

	for _, l := range listLBs.Lbs {
		err := lbAPI.DeleteLb(&lb.DeleteLbRequest{
			LbID:      l.ID,
			ReleaseIP: true,
		})
		if err != nil {
			return fmt.Errorf("error deleting lb in sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayLbBeta(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayLbBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "lb-s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbBetaExists("scaleway_lb_beta.lb01"),
					resource.TestCheckResourceAttr("scaleway_lb_beta.lb01", "name", "test-lb"),
					testCheckResourceAttrUUID("scaleway_lb_beta.lb01", "ip_id"),
					testCheckResourceAttrIPv4("scaleway_lb_beta.lb01", "ip_address"),
				),
			},
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "lb-s"
						tags = ["tag1", "tag2"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbBetaExists("scaleway_lb_beta.lb01"),
					resource.TestCheckResourceAttr("scaleway_lb_beta.lb01", "name", "test-lb"),
					resource.TestCheckResourceAttr("scaleway_lb_beta.lb01", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_lb_beta.lb01", "tags.1", "tag2"),
					testCheckResourceAttrUUID("scaleway_lb_beta.lb01", "ip_id"),
					testCheckResourceAttrIPv4("scaleway_lb_beta.lb01", "ip_address"),
				),
			},
		},
	})
}

func TestAccScalewayLbAndIPBeta(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayLbBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {
					}

					resource scaleway_lb_beta lb01 {
					    ip_id = scaleway_lb_ip_beta.ip01.id
						name = "test-lb"
						type = "lb-s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbBetaExists("scaleway_lb_beta.lb01"),
					testAccCheckScalewayLbIPBetaExists("scaleway_lb_ip_beta.ip01"),
					resource.TestCheckResourceAttr("scaleway_lb_beta.lb01", "name", "test-lb"),
					testCheckResourceAttrUUID("scaleway_lb_beta.lb01", "ip_id"),
					testCheckResourceAttrIPv4("scaleway_lb_beta.lb01", "ip_address"),
					resource.TestCheckResourceAttrPair("scaleway_lb_beta.lb01", "ip_id", "scaleway_lb_ip_beta.ip01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {
						reverse = "reverse.com" # force the refresh of the IP
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPBetaExists("scaleway_lb_ip_beta.ip01"),
					resource.TestCheckResourceAttr("scaleway_lb_ip_beta.ip01", "lb_id", ""),
				),
			},
		},
	})
}

func testAccCheckScalewayLbBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, region, ID, err := lbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetLb(&lb.GetLbRequest{
			LbID:   ID,
			Region: region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayLbBetaDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_lb_beta" {
			continue
		}

		lbAPI, region, ID, err := lbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetLb(&lb.GetLbRequest{
			Region: region,
			LbID:   ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("Load Balancer (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}
