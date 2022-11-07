package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
)

func TestAccScalewayLbFrontend_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbFrontendDestroy(tt),
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
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbFrontendExists(tt, "scaleway_lb_frontend.frt01"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend.frt01", "inbound_port", "80"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend.frt01", "timeout_client", ""),
				),
			},
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
						name = "tf-test"
						inbound_port = 443
						timeout_client = "30s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbFrontendExists(tt, "scaleway_lb_frontend.frt01"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend.frt01", "name", "tf-test"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend.frt01", "inbound_port", "443"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend.frt01", "timeout_client", "30s"),
				),
			},
		},
	})
}

// TODO: Refactor this test to enable testing of several custom domain names
//
// Let's encrypt currently has a limit of 50 certificates per week and a limit of 5 certificates per week per set of domains (including alternative names).
// So we need to change the list of alternative domain names to be able to test more than one domain name.
// One possible way to circumvent this limitation is to generate for a random set of alternative domain names that are all subdomains of the main test domain.
// For instance: *.test.scaleway-terraform.com which is a wildcard domain name.
// And we generate certificate for foo.test.scaleway-terraform.com, bar.test.scaleway-terraform.com, baz.test.scaleway-terraform.com, etc.
// Even changing one alternative domain name is enough to count as a new certificate (which is rate limited by the 50 certificates per week limit and not the 5 duplicate certificates per week limit).
// The only limitation is that all subdomains must resolve to the same IP address.
func TestAccScalewayLbFrontend_Certificate(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbFrontendDestroy(tt),
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
						forward_port = 443
						proxy_protocol = "none"
					}

					resource scaleway_lb_certificate cert01 {
						lb_id = scaleway_lb.lb01.id
						name = "test-cert-front-end"
					  	letsencrypt {
							common_name = "${replace(scaleway_lb_ip.ip01.ip_address,".", "-")}.lb.${scaleway_lb.lb01.region}.scw.cloud"
					  	}
					}

					resource scaleway_lb_frontend frt01 {
						lb_id = scaleway_lb.lb01.id
						backend_id = scaleway_lb_backend.bkd01.id
						inbound_port = 443
						certificate_ids = [scaleway_lb_certificate.cert01.id]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbFrontendExists(tt, "scaleway_lb_frontend.frt01"),
					testAccCheckScalewayFrontendCertificateExist(tt, "scaleway_lb_frontend.frt01", "scaleway_lb_certificate.cert01"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend.frt01", "certificate_ids.#", "1"),
				),
			},
		},
	})
}

func testAccCheckScalewayFrontendCertificateExist(tt *TestTools, f, c string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[f]
		if !ok {
			return fmt.Errorf("resource not found: %s", f)
		}

		cs, ok := s.RootModule().Resources[c]
		if !ok {
			return fmt.Errorf("resource not found: %s", c)
		}

		lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		frEnd, err := lbAPI.GetFrontend(&lbSDK.ZonedAPIGetFrontendRequest{
			FrontendID: ID,
			Zone:       zone,
		})
		if err != nil {
			return fmt.Errorf("error getting frontend %s: %q", rs.Primary.ID, err)
		}

		for _, id := range frEnd.CertificateIDs {
			if expandID(cs.Primary.ID) == id {
				return nil
			}
		}

		return fmt.Errorf("certificate not found: %s", c)
	}
}

func testAccCheckScalewayLbFrontendExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetFrontend(&lbSDK.ZonedAPIGetFrontendRequest{
			FrontendID: ID,
			Zone:       zone,
		})

		if err != nil {
			return fmt.Errorf("frontend not found: %s", n)
		}

		return nil
	}
}

func testAccCheckScalewayLbFrontendDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_lb_frontend" {
				continue
			}

			lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = lbAPI.GetFrontend(&lbSDK.ZonedAPIGetFrontendRequest{
				Zone:       zone,
				FrontendID: ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("LB Frontend (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return fmt.Errorf("unexpected error reading LB Frontend (%s): %s", rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func TestAclEqual(t *testing.T) {
	aclA := &lbSDK.ACL{
		Name: "test-acl",
		Match: &lbSDK.ACLMatch{
			IPSubnet:        scw.StringSlicePtr([]string{"192.168.0.1", "192.168.0.2", "192.168.10.0/24"}),
			HTTPFilter:      lbSDK.ACLHTTPFilterACLHTTPFilterNone,
			HTTPFilterValue: nil,
			Invert:          true,
		},
		Action:   &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow},
		Frontend: nil,
		Index:    1,
	}
	aclB := &lbSDK.ACL{
		Name: "test-acl",
		Match: &lbSDK.ACLMatch{
			IPSubnet:        scw.StringSlicePtr([]string{"192.168.0.1", "192.168.0.2", "192.168.10.0/24"}),
			HTTPFilter:      lbSDK.ACLHTTPFilterACLHTTPFilterNone,
			HTTPFilterValue: nil,
			Invert:          true,
		},
		Action:   &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow},
		Frontend: nil,
		Index:    1,
	}
	assert.True(t, aclEquals(aclA, aclB))

	// change name
	aclA.Name = "nope"
	assert.False(t, aclEquals(aclA, aclB))
	aclA.Name = aclB.Name

	// check action
	aclA.Action = nil
	assert.False(t, aclEquals(aclA, aclB))
	aclA.Action = &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow}
	assert.True(t, aclEquals(aclA, aclB))
	aclA.Action = &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeDeny}
	assert.False(t, aclEquals(aclA, aclB))
	aclA.Action = &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow}
	assert.True(t, aclEquals(aclA, aclB))

	// check match
	aclA.Match.IPSubnet = scw.StringSlicePtr([]string{"192.168.0.1", "192.168.0.2", "192.168.10.0/24", "0.0.0.0"})
	assert.False(t, aclEquals(aclA, aclB))
}
