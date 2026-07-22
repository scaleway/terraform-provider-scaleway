package lb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb"
	lbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb/testfuncs"
)

func TestAccCertificate_Basic(t *testing.T) {
	/**
	* See the discussion on https://github.com/scaleway/terraform-provider-scaleway/pull/396
	* Long story short, scaleway API will not permit you to request a certificate in case common name is not pointed
	* to the load balancer IP (which is unknown before creating it). In production, this can be overcome by introducing
	* an additional step which creates a DNS record and depending on it
	* We use a DNS name like: 195-154-70-235.lb.fr-par.scw.cloud
	 */
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isLbDestroyed(tt),
			lbchecks.IsIPDestroyed(tt),
			isCertificateDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}
			
					resource scaleway_lb lb01 {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb"
						type = "lb-s"
					}
			
					resource scaleway_lb_certificate cert01 {
						lb_id = scaleway_lb.lb01.id
						name = "test-cert"
					  	letsencrypt {
							common_name = "${replace(scaleway_lb_ip.ip01.ip_address,".", "-")}.lb.${scaleway_lb.lb01.region}.scw.cloud"
					  	}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_certificate.cert01", "name", "test-cert"),
					resource.TestCheckResourceAttr("scaleway_lb_certificate.cert01", "letsencrypt.#", "1"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}
			
					resource scaleway_lb lb01 {
						ip_id  = scaleway_lb_ip.ip01.id
						name   = "test-lb"
						type   = "lb-s"
					}
			
					resource scaleway_lb_certificate cert01 {
						lb_id = scaleway_lb.lb01.id
						name = "test-cert-new"
						letsencrypt {
							common_name = "${replace(scaleway_lb.lb01.ip_address, ".", "-")}.lb.${scaleway_lb.lb01.region}.scw.cloud"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_certificate.cert01", "name", "test-cert-new"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}

					resource scaleway_lb lb01 {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb"
						type = "lb-s"
					}

					resource scaleway_lb_certificate cert01 {
						lb_id = scaleway_lb.lb01.id
						name = "test-cert"
						letsencrypt {
							common_name = "${replace(scaleway_lb.lb01.ip_address, ".", "-")}.lb.${scaleway_lb.lb01.region}.scw.cloud"
							subject_alternative_name = [
							  "sub1.${replace(scaleway_lb.lb01.ip_address, ".", "-")}.lb.${scaleway_lb.lb01.region}.scw.cloud",
							  "sub2.${replace(scaleway_lb.lb01.ip_address, ".", "-")}.lb.${scaleway_lb.lb01.region}.scw.cloud"
							]
						}
					}
							`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_certificate.cert01", "name", "test-cert"),
					resource.TestCheckResourceAttr("scaleway_lb_certificate.cert01", "letsencrypt.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb_certificate.cert01", "letsencrypt.0.subject_alternative_name.#", "2"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}

					resource scaleway_lb lb01 {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb"
						type = "lb-s"
					}

					resource scaleway_lb_certificate cert01 {
						lb_id = scaleway_lb.lb01.id
						name = "test-custom-cert"
					  	custom_certificate {
							certificate_chain = <<EOF
-----BEGIN PRIVATE KEY-----
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
pWYMlA/qqGymx++DTjy1lXll6JLGPw==
-----END PRIVATE KEY-----
-----BEGIN CERTIFICATE-----
MIIF1jCCA76gAwIBAgIUXJFt07qRMnPyDndJgG/PuuH7aKowDQYJKoZIhvcNAQEL
BQAwbjELMAkGA1UEBhMCRlIxFjAUBgNVBAgMDUlsZS1kZS1GcmFuY2UxDjAMBgNV
BAcMBVBhcmlzMR0wGwYDVQQKDBRNeUludGVybmV0Q29tcGFueUxURDEYMBYGA1UE
AwwPd3d3LmV4YW1wbGUuY29tMB4XDTI2MDcyMTEwMzMwNloXDTM2MDcxODEwMzMw
NlowbjELMAkGA1UEBhMCRlIxFjAUBgNVBAgMDUlsZS1kZS1GcmFuY2UxDjAMBgNV
BAcMBVBhcmlzMR0wGwYDVQQKDBRNeUludGVybmV0Q29tcGFueUxURDEYMBYGA1UE
AwwPd3d3LmV4YW1wbGUuY29tMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKC
AgEAqzpobDbIw6WTwf52HQwCmOprvqD/nvCERjOHIqNZcbpqnP1zsa5b7INqCVED
rmG8mhZYcC5mSTv5XFiDqJYRt5xZ29q8z9yy20SNWQyJNyZd0tRvJS70SzZfbAHt
354GNUMCdyah9XhrHm2yTIq4MukmD4fOACby3g6OPIqY3OQ2z9m62voQelDDfRTZ
9qDnELHqfhVNUGCTsTXJoLpd3/LB3gVbBbQYNGWiCCfuMAey/8MnJC1mBI/QMCpb
gyUXaJk+cX2otRARAlCEVP9muxPlX33UKYe8Gv/A/5udfrflVZlc0feCAmbjUKmc
iQilqvZ7Era8VN/RcGFcHwqaP+DI2GSXW0ccPsvRZUvq+WL7Ifsi+qfeDMre2t2+
QVsGU+N4rj8ZPK8WVenRguJdvgr0qgiqSn3Z5tOOZ4boTWqgbKxWBuylOKr0wAK9
jLnsOsMCpQAJJx7XscI5wGysGl6MXaHpn9u9VER6P28ZJv3QGGPQ9uPwex3H3InJ
oHkrnesa2+NVh6pighKvT6i8kPndjr1zpYXk8UOc18lrB96WYiwWjby/Wj6+PC6I
2WNnLJPcMBqqOqMA1fOtA6DOcSk9EA2F60szR2gfcYdPfPdGUT4wWTf9Hu2BhoXJ
/YX9xn4N6pdhaePK0JIaSd/n6MUeZ6GV0GlT5bxpqYeGPpUCAwEAAaNsMGowSQYD
VR0RBEIwQIcECmQAAYcEwKgAAYIVbXlzZXJ2ZXIubXlkb21haW4uY29tghtvdGhl
cnNlcnZlci5vdGhlcmRvbWFpbi5jb20wHQYDVR0OBBYEFK1Jtpqe6HnqBmGQjnQY
q1aTGhreMA0GCSqGSIb3DQEBCwUAA4ICAQCoc/TR8lpaYxheym9BZ58boYobo/q0
LafhMSw5WkGLddAOzPx6JWNdZJptPCuQ2vmDxna6LLV6Dpxa1ZqkVf17ur9Fgz1Q
GZkxaODgl3BLFxwLAFrj3ZJY9bCnHLAt1ULjctjtkkmdYUQtPZ06VsMRQT/biGd4
qKUJ/Vj1zE99C5SpMNz4qn0qW2f+GZosI2wWhoWBmJvBZu3L4osuSwa0IpJ7i4rP
dpzzPnnO5dEva67mr2f8LrOxHF9t1+I8oGeJUPTN2bJIvwlum8V1PimyA7Uu//hR
7LzeDt1jAGxj1ufj1uDM73jb97OLkgPNHANahAELzho1HuBA5NKd9WaTq6XmiEVH
++4xgvF02GY+76jnFaNye9w0XODg2pXrhqj3RuGvEITYTT3xUaZba9KEJACmy9R2
BughLbj5kIS//sh6mMiSH9UKNDsuHXs2+lbiJVYvrsLTR6hwsur1VYw2jJnsjf2C
1/l+wmcbuvesrnWVncJy70uduXGeOh5hP+vWvCy6cjfiMbCBkA+WCQtLHL3Wf3/R
vYRO0N3UTUCfKQcI3iLKpxzoa2TYBZhU7POmidXyk6zVob7A080QYU4X+86bmLqI
eoAmUzjIyGE5covEa1J6atzAAB7arfXGSJecWeCBX8mpQ/CG66zShW7A+qWLPsi+
TKq6hPzgqta85Q==
-----END CERTIFICATE-----
EOF
					  	}
					} 
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_certificate.cert01", "name", "test-custom-cert"),
					resource.TestCheckResourceAttr("scaleway_lb_certificate.cert01", "custom_certificate.#", "1"),
				),
			},
			{
				ResourceName:            "scaleway_lb_certificate.cert01",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"custom_certificate.#", "custom_certificate.0.%", "custom_certificate.0.certificate_chain"},
			},
		},
	})
}

func isCertificateDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_lb_certificate" {
				continue
			}

			lbAPI, zone, ID, err := lb.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = lbAPI.GetCertificate(&lbSDK.ZonedAPIGetCertificateRequest{
				CertificateID: ID,
				Zone:          zone,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("LB Certificate (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
