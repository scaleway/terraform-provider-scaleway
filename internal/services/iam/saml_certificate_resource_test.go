package iam_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
)

const testSamlCertificateContent = `-----BEGIN CERTIFICATE-----
MIIDFzCCAf+gAwIBAgIURks1HTvIEoDTEmWNep/iAz4KZr0wDQYJKoZIhvcNAQEL
BQAwGzEZMBcGA1UEAwwQdGVzdC5leGFtcGxlLmNvbTAeFw0yNjA5MDgxMDI1Mjla
Fw0zNjA5MDUxMDI1MjlaMBsxGTAXBgNVBAMMEHRlc3QuZXhhbXBsZS5jb20wggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCu2UyEwLGJ1hJibBxXepoN7uMq
5N5NYEcQLawxOersENXWQEJrmBx5dChlsA8FuiM07U8yNd1qpF1JvFeAJlygCTmm
qFe2MvHraJkih49xPJxqM8HmDLuCs+XT9s2r61ciPGcJTzX2kqq44ot4RIHPQqNL
w8KfGYULYN4ypA8gnGWjLGTxvCTJNu0D0UepOOlrWTZ1rCABAAwjDEvFklEPMhDp
d3YoPJpNKTdZ4h9capPnkOGOrMYhPNHKpB9Ew2uu5Ubww5m1Zv9bEBOBE53DTtga
4geVVDux9rS1POe1gWlGOtA/Abky9Hf/XlzPAENY6Oab/PWNZPdnROFlyEmjAgMB
AAGjUzBRMB0GA1UdDgQWBBQLwBfIv5n17IQuhgFuC71ATiBPMTAfBgNVHSMEGDAW
gBQLwBfIv5n17IQuhgFuC71ATiBPMTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3
DQEBCwUAA4IBAQAgzryG5Kw55+EtnC4T3WEfXbKyHOxlZbGVGaBadP3ScnQbHdbR
I4+g5vgwe1W3UHn+cb07iwCawipi7UNw+zvN1s54W5M/t0pMM1y6xM718gqkvoIY
mNyNYBTHXVUFjlG5LvbtbU4qg6exBAQeY9dJSM9A4IYpIfjVoz9IAiILEbe3cT4d
MXRcOCSPUA6Glp5ecvRWKJsg9JnhOZXXkXWHyHw7Hj/263tRSpx/fOiIhwjH29ou
oOACpn20hAwDQOvvSXcMEtXOxCa9Itc7DzMkn28rtOhgzvkChTXG/tAGDhWcd86p
NMsLZEA3RIR7h0HU/U3u84sjYo8R59AZNpZv
-----END CERTIFICATE-----
`

func TestAccSamlCertificateResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			checkSamlCertificateDestroyed(tt),
			checkSamlDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_iam_saml" "test" {
					organization_id = "%[1]s"
				}

				resource "scaleway_iam_saml_certificate" "main" {
					saml_id = scaleway_iam_saml.test.id
					type = "signing"
					content = <<EOT
%[2]sEOT
					organization_id = "%[1]s"
					depends_on = [scaleway_iam_saml.test]
				}
			`, orgID, testSamlCertificateContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlCertificateResourceExists(tt, "scaleway_iam_saml_certificate.main"),
					resource.TestCheckResourceAttrPair("scaleway_iam_saml_certificate.main", "saml_id", "scaleway_iam_saml.test", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "type", "signing"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "content", testSamlCertificateContent),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "organization_id", orgID),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_certificate.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_certificate.main", "origin"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_certificate.main", "expires_at"),
					resource.TestMatchResourceAttr("scaleway_iam_saml_certificate.main", "srn", regexp.MustCompile(`^srn://iam\..+/saml-certificates/.+$`)),
				),
			},
			{
				ResourceName:      "scaleway_iam_saml_certificate.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSamlCertificateResource_WithDefaultOrganizationID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	_, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			checkSamlCertificateDestroyed(tt),
			checkSamlDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_saml" "test" {
					}

					resource "scaleway_iam_saml_certificate" "main" {
						saml_id = scaleway_iam_saml.test.id
						type = "signing"
						content = <<EOT
%sEOT
					}
				`, testSamlCertificateContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlCertificateResourceExists(tt, "scaleway_iam_saml_certificate.main"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "type", "signing"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "content", testSamlCertificateContent),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "origin", "identity_provider"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_certificate.main", "organization_id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_certificate.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_certificate.main", "expires_at"),
				),
			},
		},
	})
}

func checkSamlCertificateDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_iam_saml_certificate" {
				continue
			}

			samlID := rs.Primary.Attributes["saml_id"]
			if samlID == "" {
				continue
			}

			iamAPI := iam.NewAPI(tt.Meta)

			_, err := iamAPI.ListSamlCertificates(&iamSDK.ListSamlCertificatesRequest{
				SamlID: samlID,
			})
			if err == nil {
				certificates, listErr := iamAPI.ListSamlCertificates(&iamSDK.ListSamlCertificatesRequest{
					SamlID: samlID,
				})
				if listErr != nil {
					return listErr
				}

				for _, cert := range certificates.Certificates {
					if cert.ID == rs.Primary.ID {
						return fmt.Errorf("SAML certificate (%s) still exists", rs.Primary.ID)
					}
				}

				continue
			}

			if httperrors.Is404(err) {
				continue
			}

			return err
		}

		return nil
	}
}

func testAccCheckSamlCertificateResourceExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iamAPI := iam.NewAPI(tt.Meta)

		certificates, err := iamAPI.ListSamlCertificates(&iamSDK.ListSamlCertificatesRequest{
			SamlID: rs.Primary.Attributes["saml_id"],
		})
		if err != nil {
			return err
		}

		var foundCert *iamSDK.SamlCertificate

		for _, cert := range certificates.Certificates {
			if cert.ID == rs.Primary.ID {
				foundCert = cert

				break
			}
		}

		if foundCert == nil {
			return fmt.Errorf("SAML certificate (%s) not found", rs.Primary.ID)
		}

		if string(foundCert.Type) != rs.Primary.Attributes["type"] {
			return fmt.Errorf("SAML certificate type mismatch: expected %s, got %s",
				rs.Primary.Attributes["type"], foundCert.Type)
		}

		if foundCert.Content != rs.Primary.Attributes["content"] {
			return fmt.Errorf("SAML certificate content mismatch: expected %s, got %s",
				rs.Primary.Attributes["content"], foundCert.Content)
		}

		return nil
	}
}
