package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewayLbCertificateBeta(t *testing.T) {
	/**
	* Note regarding the usage of xip.io
	* See the discussion on https://github.com/terraform-providers/terraform-provider-scaleway/pull/396
	* Long story short, scaleway API will not permit you to request a certificate in case common name is not pointed
	* to the load balancer IP (which is unknown before creating it). In production, this can be overcome by introducing
	* an additional step which creates a DNS record and depending on it, but for test purposes, xip.io is an ideal solution.
	 */
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
					resource scaleway_lb_certificate_beta cert01 {
						lb_id = scaleway_lb_beta.lb01.id
						name = "test-cert"
					  	letsencrypt {
							common_name = "${scaleway_lb_beta.lb01.ip_address}.xip.io"
					  	}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_certificate_beta.cert01", "name", "test-cert"),
					resource.TestCheckResourceAttr("scaleway_lb_certificate_beta.cert01", "letsencrypt.#", "1"),
				),
			},
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "lb-s"
					}
					resource scaleway_lb_certificate_beta cert01 {
						lb_id = scaleway_lb_beta.lb01.id
						name = "test-cert-new"
					  	letsencrypt {
							common_name = "${scaleway_lb_beta.lb01.ip_address}.xip.io"
					  	}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_certificate_beta.cert01", "name", "test-cert-new"),
				),
			},
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "lb-s"
					}
					resource scaleway_lb_certificate_beta cert01 {
						lb_id = scaleway_lb_beta.lb01.id
						name = "test-cert"
					  	letsencrypt {
							common_name = "${scaleway_lb_beta.lb01.ip_address}.xip.io"
							subject_alternative_name = [
							  "sub1.${scaleway_lb_beta.lb01.ip_address}.xip.io",
							  "sub2.${scaleway_lb_beta.lb01.ip_address}.xip.io"
							]
					  	}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_certificate_beta.cert01", "name", "test-cert"),
					resource.TestCheckResourceAttr("scaleway_lb_certificate_beta.cert01", "letsencrypt.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb_certificate_beta.cert01", "letsencrypt.0.subject_alternative_name.#", "2"),
				),
			},
			{
				Config: `
					resource scaleway_lb_beta lb01 {
						name = "test-lb"
						type = "lb-s"
					}
					resource scaleway_lb_certificate_beta cert01 {
						lb_id = scaleway_lb_beta.lb01.id
						name = "test-custom-cert"
					  	custom_certificate {
							certificate_chain = <<EOF
-----BEGIN CERTIFICATE REQUEST-----
MIIBZTCBzwIBADAmMQwwCgYDVQQKDANCYXIxFjAUBgNVBAMMDSouZXhhbXBsZS5j
b20wgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAKM8iE+6xpL9ps2x/W/+0wf3
+5Kgcwl+AoqUTNmAT6eaLfj3PMtdAmN+IcGsYXK2F1FENAdfHd70m4mSBoMJeq+F
a2ZEYp+RlFQqldTBSUBld48z9DNkbNmb0cMyVmntnKhPZ5wjNjgsYee2Leesm2lx
Gw2BCpDxGLBmwKFca3bTAgMBAAGgADANBgkqhkiG9w0BAQsFAAOBgQBbelf1jYNX
Y+AbCpbsySJ32iDtlrZkcIDHMlRIefnoXnO8B1RiEY0SU/n/b2tFCAcvkhq0whKj
RPOH6lzT0x3lWxI+C6xpVyURlljHPygVRbD3fig1rFnH3sI2IE3fJEbGJdmmHacA
BRgm3RR+HWWLKB0ygUbOtZk0BLBUqPK+WQ==
-----END CERTIFICATE REQUEST-----
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCjPIhPusaS/abNsf1v/tMH9/uSoHMJfgKKlEzZgE+nmi349zzL
XQJjfiHBrGFythdRRDQHXx3e9JuJkgaDCXqvhWtmRGKfkZRUKpXUwUlAZXePM/Qz
ZGzZm9HDMlZp7ZyoT2ecIzY4LGHnti3nrJtpcRsNgQqQ8RiwZsChXGt20wIDAQAB
AoGBAINGbhU4jvOtW9T2fGvyEgLJkp7zvC/5D9Akvbz5LJYML0aWhmTB0ubyi/E2
UVQwToZDhFgdTWd9bgxvzB7bo7Z1kiKWTXQSUDTjC3fchE12UU+1/s0LGh9B1cpV
YqFAu7QNG4maXdZ2h+u+T7dioHslS74PFCBExClkwDiUsdcRAkEAzufrw6zgC0R3
x6OMaaPsH2oX9KfR0+Kxj2m02TnwyFPSrSlIGfoSVtU5wBkvCwTBx/IPY70vNw3f
fPacJNldGwJBAMn3/2nqMJUhR3U7bZA3lI+T1z01D7dLbvCafh3oHTYRAbHO2f0z
k1uhn1+2r9QICwaugZwvx7biCfT3rTqXAKkCQGbINwphWnq+bHI0AJCJ6cZBQd07
cLS9LE99x2URr1cUrNdwZmzhGTMhgSq4V/I1Tr4wtQxq8oV60saVC0QS5nkCQQDI
W2NfuNlVN9xhqgC4zspr3KfrqlXa6dQ2j6yJEpjX5+scby3Fh4Kpph4qn1qyJwB5
MmiVfrjK7lYeVA3fT6lxAkAc1QlmGapJ3w9996v6OZ3UqNBzsHsMmze/zGemdMfc
VaSSB5OL7dZ8fLoJctv9EwgIBM3LFhSgAm9ASPHHR9fp
-----END RSA PRIVATE KEY-----
----BEGIN TRUSTED CERTIFICATE-----
MIIBmTCCAQICCQCb2NiBTIlkeTANBgkqhkiG9w0BAQUFADARMQ8wDQYDVQQKDAZG
b29CYXIwHhcNMjAwMzAzMTYwODQ4WhcNMzAwMzAxMTYwODQ4WjARMQ8wDQYDVQQK
DAZGb29CYXIwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALba1p/3sJ3E2J0J
Wa6N4Zw9aBAq4OedYszUCYqviG6XiCKZPeI1skfD9bef4pYG67R5OcccCsjwqfT4
HkwKj/2SG4VsJZpvVr9S1uOeiA70KywB9GvuiLZUflrDEy4IEA0s6UBALRgXllX+
QbnE8VC9brzgepQwO6tTU1Vjo9s9AgMBAAEwDQYJKoZIhvcNAQEFBQADgYEAelXe
TxdxZWXZFldVl7LUS39lsc9CRgIcLbmk/s23zs92zyJcYvarulx2ku4Zrp507RzT
hPrE4kljWeyQnwIXeOSg5UZrsRtRxUTDLBtOrL3PmFZhBaDShSc+1DS4U/kjy/k+
F+6t5ymALbasChWwC7FFgSsF5PtGP4PxZHFdXiE=
-----END TRUSTED CERTIFICATE-----
EOF
					  	}
					} 
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_lb_certificate_beta.cert01", "name", "test-custom-cert"),
					resource.TestCheckResourceAttr("scaleway_lb_certificate_beta.cert01", "custom_certificate.#", "1"),
				),
			},
		},
	})
}
