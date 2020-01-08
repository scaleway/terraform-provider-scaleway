package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewayLbAclBeta(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayLbAclBetaDestroy,
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
					resource scaleway_lb_acl_beta acl01 {
						frontend_id = scaleway_lb_frontend_beta.frt01.id
						name = "test-acl"
						action {
							type = "allow"
						}
						match {
							ip_subnet = ["192.168.0.1", "192.168.0.2", "192.168.10.0/24"]
							http_filter = "acl_http_filter_none"
							http_filter_value = ["criteria1","criteria2"]
							invert = "true"
						}
						index = "42"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbAclBetaExists("scaleway_lb_acl_beta.acl01"),
					resource.TestCheckResourceAttr("scaleway_lb_acl_beta.acl01", "index", "42"),
					resource.TestCheckResourceAttr("scaleway_lb_acl_beta.acl01", "match.0.http_filter", "acl_http_filter_none"),
				),
			},
		},
	})
}

func testAccCheckScalewayLbAclBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, region, ID, err := getLbAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetACL(&lb.GetACLRequest{
			Region: region,
			ACLID:  ID,
		})

		if err != nil {
			return err
		}

		return nil
	}
}
func testAccCheckScalewayLbAclBetaDestroy(s *terraform.State) error {
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
