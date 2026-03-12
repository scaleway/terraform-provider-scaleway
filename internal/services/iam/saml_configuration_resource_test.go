package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccSamlConfigurationResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             checkSamlDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_saml" "main" {
						organization_id = "%s"
					}

					resource "scaleway_iam_saml_configuration" "main" {
						organization_id = scaleway_iam_saml.main.organization_id
						entity_id = "https://example.com/saml/metadata"
						single_sign_on_url = "https://example.com/saml/sso"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlResourceExists(tt, "scaleway_iam_saml_configuration.main"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "organization_id", orgID),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "entity_id", "https://example.com/saml/metadata"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "single_sign_on_url", "https://example.com/saml/sso"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "status", "missing_certificate"),
					resource.TestCheckResourceAttrPair("scaleway_iam_saml_configuration.main", "id", "scaleway_iam_saml.main", "id"),
				),
			},
			{
				ResourceName:      "scaleway_iam_saml_configuration.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSamlConfigurationResource_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             checkSamlDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_saml" "main" {
						organization_id = "%s"
					}

					resource "scaleway_iam_saml_configuration" "main" {
						organization_id = scaleway_iam_saml.main.organization_id
						entity_id = "https://example.com/saml/metadata"
						single_sign_on_url = "https://example.com/saml/sso"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlResourceExists(tt, "scaleway_iam_saml_configuration.main"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "entity_id", "https://example.com/saml/metadata"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "single_sign_on_url", "https://example.com/saml/sso"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_saml" "main" {
						organization_id = "%s"
					}

					resource "scaleway_iam_saml_configuration" "main" {
						organization_id = scaleway_iam_saml.main.organization_id
						entity_id = "https://updated-example.com/saml/metadata"
						single_sign_on_url = "https://updated-example.com/saml/sso"
						depends_on = [scaleway_iam_saml.main]
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlResourceExists(tt, "scaleway_iam_saml_configuration.main"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "entity_id", "https://updated-example.com/saml/metadata"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "single_sign_on_url", "https://updated-example.com/saml/sso"),
				),
			},
		},
	})
}

func TestAccSamlConfigurationResource_WithDefaultOrganizationID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	_, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             checkSamlDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_saml" "main" {
					}

					resource "scaleway_iam_saml_configuration" "main" {
						entity_id = "https://example.com/saml/metadata"
						single_sign_on_url = "https://example.com/saml/sso"
						depends_on = [scaleway_iam_saml.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlResourceExists(tt, "scaleway_iam_saml_configuration.main"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "entity_id", "https://example.com/saml/metadata"),
					resource.TestCheckResourceAttr("scaleway_iam_saml_configuration.main", "single_sign_on_url", "https://example.com/saml/sso"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_configuration.main", "organization_id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml_configuration.main", "id"),
				),
			},
		},
	})
}
