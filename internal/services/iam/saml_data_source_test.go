package iam_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceSaml_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             checkSamlDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_saml" "main" {
						organization_id = "%s"
					}

					data "scaleway_iam_saml" "main" {
						organization_id = scaleway_iam_saml.main.organization_id
						depends_on = [scaleway_iam_saml.main]
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "organization_id", "scaleway_iam_saml.main", "organization_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "id", "scaleway_iam_saml.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "status", "scaleway_iam_saml.main", "status"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "service_provider.entity_id", "scaleway_iam_saml.main", "service_provider.entity_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "service_provider.assertion_consumer_service_url", "scaleway_iam_saml.main", "service_provider.assertion_consumer_service_url"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "entity_id", "scaleway_iam_saml.main", "entity_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "single_sign_on_url", "scaleway_iam_saml.main", "single_sign_on_url"),
				),
			},
		},
	})
}

func TestAccDataSourceSaml_WithDefaultOrganizationID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	_, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}
	{
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: tt.ProviderFactories,
			CheckDestroy:             checkSamlDestroyed(tt),
			Steps: []resource.TestStep{
				{
					Config: `
					resource "scaleway_iam_saml" "main" {
					}

					data "scaleway_iam_saml" "main" {
						depends_on = [scaleway_iam_saml.main]
					}
				`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "organization_id", "scaleway_iam_saml.main", "organization_id"),
						resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "id", "scaleway_iam_saml.main", "id"),
						resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "status", "scaleway_iam_saml.main", "status"),
						resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "service_provider.entity_id", "scaleway_iam_saml.main", "service_provider.entity_id"),
						resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "service_provider.assertion_consumer_service_url", "scaleway_iam_saml.main", "service_provider.assertion_consumer_service_url"),
						resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "entity_id", "scaleway_iam_saml.main", "entity_id"),
						resource.TestCheckResourceAttrPair("data.scaleway_iam_saml.main", "single_sign_on_url", "scaleway_iam_saml.main", "single_sign_on_url"),
					),
				},
			},
		})
	}
}

func TestAccDataSourceSaml_InvalidDeactivated(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_iam_saml" "main" {
						organization_id = "%s"
					}
				`, orgID),
				ExpectError: regexp.MustCompile("not found"),
			},
		},
	})
}
