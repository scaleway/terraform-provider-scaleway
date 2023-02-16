package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

func TestAccScalewayLbRoute_WithSNI(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbRouteDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip ip01 {}
					resource scaleway_lb lb01 {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb"
						type = "lb-s"
					}
					resource scaleway_lb_backend bkd01 {
						lb_id = scaleway_lb.lb01.id
						forward_protocol = "tcp"
						forward_port = 80
						proxy_protocol = "none"
					}
					resource scaleway_lb_frontend frt01 {
						lb_id = scaleway_lb.lb01.id
						backend_id = scaleway_lb_backend.bkd01.id
						inbound_port = 80
					}
					resource scaleway_lb_route rt01 {
						frontend_id = scaleway_lb_frontend.frt01.id
						backend_id = scaleway_lb_backend.bkd01.id
						match_sni = "sni.scaleway.com"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbRouteExists(tt, "scaleway_lb_route.rt01"),
					resource.TestCheckResourceAttr("scaleway_lb_route.rt01", "match_sni", "sni.scaleway.com"),
				),
			},
		},
	})
}

func TestAccScalewayLbRoute_WithHostHeader(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbRouteDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip ip01 {}
					resource scaleway_lb lb01 {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb"
						type = "lb-s"
					}
					resource scaleway_lb_backend bkd01 {
						lb_id = scaleway_lb.lb01.id
						forward_protocol = "http"
						forward_port = 80
						proxy_protocol = "none"
					}
					resource scaleway_lb_frontend frt01 {
						lb_id = scaleway_lb.lb01.id
						backend_id = scaleway_lb_backend.bkd01.id
						inbound_port = 80
					}
					resource scaleway_lb_route rt01 {
						frontend_id = scaleway_lb_frontend.frt01.id
						backend_id = scaleway_lb_backend.bkd01.id
						match_host_header = "host.scaleway.com" 
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbRouteExists(tt, "scaleway_lb_route.rt01"),
					resource.TestCheckResourceAttr("scaleway_lb_route.rt01", "match_host_header", "host.scaleway.com"),
				),
			},
		},
	})
}

func testAccCheckScalewayLbRouteExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetRoute(&lbSDK.ZonedAPIGetRouteRequest{
			RouteID: ID,
			Zone:    zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayLbRouteDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_lb_route" {
				continue
			}

			lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = lbAPI.GetRoute(&lbSDK.ZonedAPIGetRouteRequest{
				Zone:    zone,
				RouteID: ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("LB Route (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
