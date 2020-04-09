package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

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
