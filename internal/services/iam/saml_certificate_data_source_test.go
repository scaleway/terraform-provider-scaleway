package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceSamlCertificate_Basic(t *testing.T) {
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

					data "scaleway_iam_saml_certificate" "main" {
						certificate_id = scaleway_iam_saml_certificate.main.id
					}
				`, orgID, certContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlCertificateResourceExists(tt, "scaleway_iam_saml_certificate.main"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml_certificate.main", "certificate_id", "scaleway_iam_saml_certificate.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml_certificate.main", "content", "scaleway_iam_saml_certificate.main", "content"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml_certificate.main", "type", "scaleway_iam_saml_certificate.main", "type"),
					resource.TestCheckResourceAttrSet("data.scaleway_iam_saml_certificate.main", "origin"),
					resource.TestCheckResourceAttrSet("data.scaleway_iam_saml_certificate.main", "expires_at"),
				),
			},
		},
	})
}

func TestAccDataSourceSamlCertificate_WithDefaultOrganizationID(t *testing.T) {
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

					data "scaleway_iam_saml_certificate" "main" {
						certificate_id = scaleway_iam_saml_certificate.main.id
						depends_on = [scaleway_iam_saml_certificate.main]
					}
				`, certContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlCertificateResourceExists(tt, "scaleway_iam_saml_certificate.main"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml_certificate.main", "certificate_id", "scaleway_iam_saml_certificate.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml_certificate.main", "content", "scaleway_iam_saml_certificate.main", "content"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml_certificate.main", "type", "scaleway_iam_saml_certificate.main", "type"),
					resource.TestCheckResourceAttrSet("data.scaleway_iam_saml_certificate.main", "origin"),
					resource.TestCheckResourceAttrSet("data.scaleway_iam_saml_certificate.main", "expires_at"),
				),
			},
		},
	})
}
