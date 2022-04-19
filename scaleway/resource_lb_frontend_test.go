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

// de votre coté vous allez rester limité par LE sur le nombre de certif issue pour un domain précis malheureusement (Certificates per Registered Domain (50 per week))
// donc tu tombes sur Duplicate Certificate limit of 5 per week
//11 h 29
//pour que vous puissez déjà passer à 50 par semaine au lieu de 5 faudrait que tu random sur plein de sous-domain à ton domain de test
//11 h 30
//tu te fais un dns entry *.test.scaleway-terraform.com
//11 h 30
//et dans ton test tu random sur un chiffre x.test.scaleway-terraform.com
//11 h 30
//comme ça déjà vous pourrez run plus de test
// ça reste sur le même domain de base donc tout xxxxxx.test.scaleway-terraform.com compte pour les 50
// mais si tu fais varier le sous-domain ouais au moins c’est pas juste 5
// tu peux laisser en domain principal le test.scaleway-terraform.com juste tu rajoutes des alt domain name avec un autre au pif X.test.scaleway-terraform.com
// c’est vraiment le set de domain complet qui est check, si une chaine de 5 domain name, tu en fais varier qu’un seul c’est bon
// la seule limitation c’est qu’ils soient résolu aussi sur l’ip du challnge
func TestAccScalewayLbFrontend_Certificate(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	testDNSZone := fmt.Sprintf("test.%s", testDomain)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbFrontendDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_lb_ip ip01 {}

					resource "scaleway_domain_record" "tf_A" {
						dns_zone = %[1]q
						name     = "test"
						type     = "A"
						data     = "${scaleway_lb_ip.ip01.ip_address}"
						ttl      = 3600
						priority = 1
					}

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

					resource scaleway_lb_certificate cert02 {
						lb_id = scaleway_lb.lb01.id
						name = "test-cert-front-end2"
					  	letsencrypt {
							common_name = %[2]q
					  	}
					}

					resource scaleway_lb_frontend frt01 {
						lb_id = scaleway_lb.lb01.id
						backend_id = scaleway_lb_backend.bkd01.id
						inbound_port = 443
						certificate_ids = [scaleway_lb_certificate.cert01.id, scaleway_lb_certificate.cert02.id]
					}
				`, testDomain, testDNSZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbFrontendExists(tt, "scaleway_lb_frontend.frt01"),
					testAccCheckScalewayFrontendCertificateExist(tt, "scaleway_lb_frontend.frt01", "scaleway_lb_certificate.cert01"),
					testAccCheckScalewayFrontendCertificateExist(tt, "scaleway_lb_frontend.frt01", "scaleway_lb_certificate.cert02"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend.frt01",
						"certificate_ids.#", "2"),
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
			return err
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
			return err
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
				return err
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

	//change name
	aclA.Name = "nope"
	assert.False(t, aclEquals(aclA, aclB))
	aclA.Name = aclB.Name

	//check action
	aclA.Action = nil
	assert.False(t, aclEquals(aclA, aclB))
	aclA.Action = &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow}
	assert.True(t, aclEquals(aclA, aclB))
	aclA.Action = &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeDeny}
	assert.False(t, aclEquals(aclA, aclB))
	aclA.Action = &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow}
	assert.True(t, aclEquals(aclA, aclB))

	//check match
	aclA.Match.IPSubnet = scw.StringSlicePtr([]string{"192.168.0.1", "192.168.0.2", "192.168.10.0/24", "0.0.0.0"})
	assert.False(t, aclEquals(aclA, aclB))
}
