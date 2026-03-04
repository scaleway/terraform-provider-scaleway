package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
)

func TestAccSamlResource_Basic(t *testing.T) {
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
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlResourceExists(tt, "scaleway_iam_saml.main"),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "organization_id", orgID),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "entity_id", ""),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "single_sign_on_url", ""),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "status", "missing_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml.main", "service_provider.entity_id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml.main", "service_provider.assertion_consumer_service_url"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_saml" "main" {
						organization_id = "%s"
						entity_id = "https://example.com/saml/metadata"
						single_sign_on_url = "https://example.com/saml/sso"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlResourceExists(tt, "scaleway_iam_saml.main"),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "entity_id", "https://example.com/saml/metadata"),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "single_sign_on_url", "https://example.com/saml/sso"),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "status", "missing_certificate"),
				),
			},
			{
				ResourceName:      "scaleway_iam_saml.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSamlResource_WithDefaultOrganizationID(t *testing.T) {
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
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlResourceExists(tt, "scaleway_iam_saml.main"),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "entity_id", ""),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "single_sign_on_url", ""),
					resource.TestCheckResourceAttr("scaleway_iam_saml.main", "status", "missing_certificate"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml.main", "organization_id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_saml.main", "id"),
				),
			},
		},
	})
}

func checkSamlDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_iam_saml" {
				continue
			}

			iamAPI := iam.NewAPI(tt.Meta)

			_, err := iamAPI.GetOrganizationSaml(&iamSDK.GetOrganizationSamlRequest{
				OrganizationID: rs.Primary.Attributes["organization_id"],
			})

			if err == nil {
				return fmt.Errorf("SAML configuration (%s) still exists", rs.Primary.ID)
			}
			if httperrors.Is404(err) {
				continue
			}
			return err
		}

		return nil
	}
}

func testAccCheckSamlResourceExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iamAPI := iam.NewAPI(tt.Meta)

		saml, err := iamAPI.GetOrganizationSaml(&iamSDK.GetOrganizationSamlRequest{
			OrganizationID: rs.Primary.Attributes["organization_id"],
		})
		if err != nil {
			return err
		}

		if saml.EntityID != rs.Primary.Attributes["entity_id"] {
			return fmt.Errorf("SAML entity_id mismatch: expected %s, got %s",
				rs.Primary.Attributes["entity_id"], saml.EntityID)
		}

		if saml.SingleSignOnURL != rs.Primary.Attributes["single_sign_on_url"] {
			return fmt.Errorf("SAML single_sign_on_url mismatch: expected %s, got %s",
				rs.Primary.Attributes["single_sign_on_url"], saml.SingleSignOnURL)
		}

		return nil
	}
}
