package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	scwClient, err := sharedClientForRegion(scw.Region(region))
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	lbAPI := lb.NewAPI(scwClient)

	l.Debugf("sweeper: destroying the lbs in (%s)", region)
	listLBs, err := lbAPI.ListLBs(&lb.ListLBsRequest{}, scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("error listing lbs in (%s) in sweeper: %s", region, err)
	}

	for _, l := range listLBs.LBs {
		err := lbAPI.DeleteLB(&lb.DeleteLBRequest{
			LBID:      l.ID,
			ReleaseIP: true,
		})
		if err != nil {
			return fmt.Errorf("error deleting lb in sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayLbLb_WithIP(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {
					}

					resource scaleway_lb_beta lb01 {
					    ip_id = scaleway_lb_ip_beta.ip01.id
						name = "test-lb"
						type = "LB-S"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbBetaExists(tt, "scaleway_lb_beta.lb01"),
					testAccCheckScalewayLbIPBetaExists(tt, "scaleway_lb_ip_beta.ip01"),
					resource.TestCheckResourceAttr("scaleway_lb_beta.lb01", "name", "test-lb"),
					testCheckResourceAttrUUID("scaleway_lb_beta.lb01", "ip_id"),
					testCheckResourceAttrIPv4("scaleway_lb_beta.lb01", "ip_address"),
					resource.TestCheckResourceAttrPair("scaleway_lb_beta.lb01", "ip_id", "scaleway_lb_ip_beta.ip01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPBetaExists(tt, "scaleway_lb_ip_beta.ip01"),
				),
			},
		},
	})
}

func testAccCheckScalewayLbBetaExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, region, ID, err := lbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetLB(&lb.GetLBRequest{
			LBID:   ID,
			Region: region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayLbBetaDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_lb_beta" {
				continue
			}

			lbAPI, region, ID, err := lbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = lbAPI.GetLB(&lb.GetLBRequest{
				Region: region,
				LBID:   ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("load Balancer (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
