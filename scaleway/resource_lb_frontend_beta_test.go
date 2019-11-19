package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

func TestAccScalewayLbFrontendBeta(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayLbFrontendBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "lb-s"
					}
					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						forward_protocol = "tcp"
						forward_port = 80
					}

					resource scaleway_lb_frontend_beta frt01 {
						lb_id = scaleway_lb_beta.lb01.id
						backend_id = scaleway_lb_backend_beta.bkd01.id
						inbound_port = 80
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbFrontendBetaExists("scaleway_lb_frontend_beta.frt01"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "inbound_port", "80"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "timeout_client", ""),
				),
			},
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "lb-s"
					}
					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						forward_protocol = "tcp"
						forward_port = 80
					}
					resource scaleway_lb_frontend_beta frt01 {
						lb_id = scaleway_lb_beta.lb01.id
						backend_id = scaleway_lb_backend_beta.bkd01.id
						name = "tf-test"
						inbound_port = 443
						timeout_client = "30s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbFrontendBetaExists("scaleway_lb_frontend_beta.frt01"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "name", "tf-test"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "inbound_port", "443"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "timeout_client", "30s"),
				),
			},
		},
	})
}

func testAccCheckScalewayLbFrontendBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, region, ID, err := getLbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetFrontend(&lb.GetFrontendRequest{
			FrontendID: ID,
			Region:     region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayLbFrontendBetaDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_lb_frontend_beta" {
			continue
		}

		lbAPI, region, ID, err := getLbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetFrontend(&lb.GetFrontendRequest{
			Region:     region,
			FrontendID: ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("LB Frontend (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}
