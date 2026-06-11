package iam_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
)

func generateTestCert() (string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.example.com"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(2 * time.Minute),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return "", err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	return string(certPEM), nil
}

func TestAccSamlCertificateResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	certContent, err := generateTestCert()
	if err != nil {
		t.Error("Failed to generate test certificate")
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
			`, orgID, certContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlCertificateResourceExists(tt, "scaleway_iam_saml_certificate.main"),
					resource.TestCheckResourceAttrPair("scaleway_iam_saml_certificate.main", "saml_id", "scaleway_iam_saml.test", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "type", "signing"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "content", certContent),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "organization_id", orgID),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_certificate.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_certificate.main", "origin"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_certificate.main", "expires_at"),
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

	certContent, err := generateTestCert()
	if err != nil {
		t.Error("Failed to generate test certificate")
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
				`, certContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlCertificateResourceExists(tt, "scaleway_iam_saml_certificate.main"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "type", "signing"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_certificate.main", "content", certContent),
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
