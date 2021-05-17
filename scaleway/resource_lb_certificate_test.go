package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayLbCertificate_Basic(t *testing.T) {
	/**
	* See the discussion on https://github.com/scaleway/terraform-provider-scaleway/pull/396
	* Long story short, scaleway API will not permit you to request a certificate in case common name is not pointed
	* to the load balancer IP (which is unknown before creating it). In production, this can be overcome by introducing
	* an additional step which creates a DNS record and depending on it
	* We use a DNS name like: 195-154-70-235.lb.fr-par.scw.cloud
	 */
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbDestroy(tt),
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
-----BEGIN CERTIFICATE-----
MIIDXzCCAkegAwIBAgIUCQTOZjf11LGYmvGVw7xxWurGFgMwDQYJKoZIhvcNAQEL
BQAwSDELMAkGA1UEBhMCRlIxDjAMBgNVBAcMBVBhcmlzMREwDwYDVQQKDAhTY2Fs
ZXdhdTEWMBQGA1UEAwwNKi5leGFtcGxlLmNvbTAeFw0yMDA0MjgxNzQ2MzFaFw0y
MjA0MjgxNzQ2MzFaMEgxCzAJBgNVBAYTAkZSMQ4wDAYDVQQHDAVQYXJpczERMA8G
A1UECgwIU2NhbGV3YXUxFjAUBgNVBAMMDSouZXhhbXBsZS5jb20wggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQCws8az+MXTtmNKN66Bl2HmlkAxazMcK9eG
rmU5jOs3nNZevjn4FhQClVt2mOX7G37b32wpLyt5MUYGg+Ac95fL5zx9nwp57YBx
NteGYIUpLTD2tb3aKuuIM+eSFkRabOsN1cqwqtMK8pW1YA48olhrunxW82SrMKUw
cFN508wochbeupkk6nG3K29+1Uhwzr9B93xVQ5FQZKSvj5fuC3OIWJitpVyWlGBx
pdoTq6T5W3H0odCNoUx8U0nKjOtcPoupe4ZAHL5VEmKi/PaNn4Wpjf0e6KSQv2ld
WjCyaxSMpVpDSIsGPwy1symPV5Vk586Oexy2A5DIMlLRhGqAR9VbAgMBAAGjQTA/
MA4GA1UdDwEB/wQEAwIDiDATBgNVHSUEDDAKBggrBgEFBQcDATAYBgNVHREEETAP
gg0qLmV4YW1wbGUuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQACL1iY9Az8VQM4sOKS
NYYenvlzGalfQGrNh76DD5edv+0a5YQhWiVXoig0BIwPrsEGR1D4epH4Hrp9Uw56
5Pln7kpZmcr3rm5WNxYLVBp7uTbAuFAiZZqLKPp7n3eay5AYjGYNa0DEvm0PijF3
YqpCulrHYNFi0kxwNuoyJ6lORWjbCchxW14zB65hM1nLC4iXUtMF1cin1oOTtRKg
7o6bLtz7BSdkZE/VSpEg17KO0bR9NWraBby08sf5QcPWrzYYyeoCoR9jC3a2ifUP
XlnKjeabTGw/NxayLLRJLu5+dJuPodHm/I1Uwl1QWSbh4d1FQBX436mK41zMZhg0
hUVl
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
		},
	})
}
