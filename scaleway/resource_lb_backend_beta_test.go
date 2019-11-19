package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
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
						type = "lb-s"
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
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "timeout_server", ""),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "on_marked_down_action", "none"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_timeout", "30s"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_port", "80"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_max_retries", "2"),
					resource.TestCheckResourceAttrPair("scaleway_lb_backend_beta.bkd01", "server_ips.0", "scaleway_instance_ip.ip01", "address"),
				),
			},
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "lb-s"
					}

					resource scaleway_instance_ip ip01 {}
					resource scaleway_instance_ip ip02 {}

					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						name = "bkd01"
						forward_protocol = "tcp"
						forward_port = 80
						forward_port_algorithm = "leastconn"
						sticky_sessions = "cookie"
						sticky_sessions_cookie_name = "session-id"
						server_ips = [ scaleway_instance_ip.ip02.address ]
						send_proxy_v2 = true
						timeout_server = "1s"
						timeout_connect = "2.5s"
						timeout_tunnel = "3s"
						health_check_timeout = "15s"
						health_check_port = 81
						health_check_max_retries = 3
						on_marked_down_action = "shutdown_sessions"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbBackendBetaExists("scaleway_lb_backend_beta.bkd01"),
					resource.TestCheckResourceAttrPair("scaleway_lb_backend_beta.bkd01", "server_ips.0", "scaleway_instance_ip.ip02", "address"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "on_marked_down_action", "shutdown_sessions"),
				),
			},
		},
	})
}

func TestAccScalewayLbBackendBeta_HealthCheck(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayLbBackendBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "lb-s"
					}

					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						name = "bkd01"
						forward_protocol = "tcp"
						forward_port = 80
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_tcp.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_http.#", "0"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_https.#", "0"),
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
						name = "bkd01"
						forward_protocol = "tcp"
						forward_port = 80

						health_check_http {
							uri = "http://test.com/health"
							method = "POST"
							code = 404
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_tcp.#", "0"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_http.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_https.#", "0"),
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
						name = "bkd01"
						forward_protocol = "tcp"
						forward_port = 80

						health_check_https {
							uri = "http://test.com/health"
							method = "POST"
							code = 404
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_tcp.#", "0"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_http.#", "0"),
					resource.TestCheckResourceAttr("scaleway_lb_backend_beta.bkd01", "health_check_https.#", "1"),
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
