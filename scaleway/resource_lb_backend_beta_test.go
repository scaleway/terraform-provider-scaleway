package scaleway

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"testing"
)

func TestAccScalewayLbBackendBeta(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayLbBackendBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "LB-S"
					}

					resource scaleway_instance_ip ip01 {}
					resource scaleway_instance_ip ip02 {}

					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						name = "bkd01"
						forward_protocol = "tcp"
						forward_port = 80
						server_ips = [ scaleway_instance_ip.ip01.address ]
					}

				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbBackendBetaExists("scaleway_lb_backend_beta.bkd01"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "forward_port_algorithm", "roundrobin"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "sticky_sessions", "none"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "send_proxy_v2", "false"),
				),
			},
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "LB-S"
					}

					resource scaleway_instance_ip ip01 {}
					resource scaleway_instance_ip ip02 {}

					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						name = "bkd01"
						forward_protocol = "tcp"
						forward_port = 80
						forward_port_algorithm = "roundrobin"
						sticky_sessions = "cookie"
						sticky_sessions_cookie_name = "session-id"
						server_ips = [ scaleway_instance_ip.ip02.address ]
						send_proxy_v2 = true
						//timeout_server
						//timeout_connect
						//timeout_tunnel
						//on_marked_down_action
					}

				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbBackendBetaExists("scaleway_lb_backend_beta.bkd01"),
				),
			},
		},
	})
}

func testAccCheckScalewayLbBackendBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, region, ID, err := getLbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetBackend(&lb.GetBackendRequest{
			BackendID: ID,
			Region:    region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayLbBackendBetaDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_lb_backend_beta" {
			continue
		}

		lbAPI, region, ID, err := getLbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetBackend(&lb.GetBackendRequest{
			Region:    region,
			BackendID: ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("LB Backend (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}
