package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

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

func testAccCheckScalewayLbBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, region, ID, err := getLbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
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

		lbAPI, region, ID, err := getLbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetLb(&lb.GetLbRequest{
			Region: region,
			LbID:   ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("IP (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		// We check for 403 because instance API return 403 for deleted IP
		if !is404Error(err) {
			return err
		}
	}

	return nil
}
