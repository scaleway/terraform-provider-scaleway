package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionUpdateSamlConfiguration_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionUpdateSamlConfiguration_Basic because actions are not yet supported on OpenTofu")
	}

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
						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_iam_update_saml_configuration.main]
							}
						}
					}

					action "scaleway_iam_update_saml_configuration" "main" {
						config {
							organization_id = scaleway_iam_saml.main.organization_id
							entity_id = "https://example.com/saml/metadata"
							single_sign_on_url = "https://example.com/saml/sso"
						}
					}
				`, orgID),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_saml" "main" {
						organization_id = "%s"
						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_iam_update_saml_configuration.main]
							}
						}
					}

					action "scaleway_iam_update_saml_configuration" "main" {
						config {
							organization_id = scaleway_iam_saml.main.organization_id
							entity_id = "https://example.com/saml/metadata"
							single_sign_on_url = "https://example.com/saml/sso"
						}
					}

					data "scaleway_iam_saml" "main" {
						organization_id = scaleway_iam_saml.main.organization_id
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlResourceExists(tt, "scaleway_iam_saml.main"),
					resource.TestCheckResourceAttr("data.scaleway_iam_saml.main", "entity_id", "https://example.com/saml/metadata"),
					resource.TestCheckResourceAttr("data.scaleway_iam_saml.main", "single_sign_on_url", "https://example.com/saml/sso"),
				),
			},
		},
	})
}

func TestAccActionUpdateSamlConfiguration_WithDefaultOrganizationID(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionUpdateSamlConfiguration_WithDefaultOrganizationID because actions are not yet supported on OpenTofu")
	}

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
						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_iam_update_saml_configuration.main]
							}
						}
					}

					action "scaleway_iam_update_saml_configuration" "main" {
						config {
							entity_id = "https://example.com/saml/metadata"
							single_sign_on_url = "https://example.com/saml/sso"
						}
					}
				`,
			},
			{
				Config: `
					resource "scaleway_iam_saml" "main" {
						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_iam_update_saml_configuration.main]
							}
						}
					}

					action "scaleway_iam_update_saml_configuration" "main" {
						config {
							entity_id = "https://example.com/saml/metadata"
							single_sign_on_url = "https://example.com/saml/sso"
						}
					}

					data "scaleway_iam_saml" "main" {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_iam_saml.main", "entity_id", "https://example.com/saml/metadata"),
					resource.TestCheckResourceAttr("data.scaleway_iam_saml.main", "single_sign_on_url", "https://example.com/saml/sso"),
					resource.TestCheckResourceAttrSet("data.scaleway_iam_saml.main", "organization_id"),
				),
			},
		},
	})
}
