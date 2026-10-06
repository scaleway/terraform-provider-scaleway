package iam_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceScimToken_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			checkScimTokenDestroyed(tt),
			checkScimDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_scim" "main" {
						organization_id = "%s"
					}

					resource "scaleway_iam_scim_token" "main" {
						scim_id = scaleway_iam_scim.main.id
						organization_id = "%[1]s"
					}

					data "scaleway_iam_scim_token" "main" {
						scim_id = scaleway_iam_scim.main.id
						token_id = scaleway_iam_scim_token.main.id
						organization_id = "%[1]s"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "organization_id", "scaleway_iam_scim_token.main", "organization_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "scim_id", "scaleway_iam_scim_token.main", "scim_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "token_id", "scaleway_iam_scim_token.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "created_at", "scaleway_iam_scim_token.main", "created_at"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "expires_at", "scaleway_iam_scim_token.main", "expires_at"),
				),
			},
		},
	})
}

func TestAccDataSourceScimToken_WithDefaultOrganizationID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	_, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			checkScimTokenDestroyed(tt),
			checkScimDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_scim" "main" {
					}

					resource "scaleway_iam_scim_token" "main" {
						scim_id = scaleway_iam_scim.main.id
					}

					data "scaleway_iam_scim_token" "main" {
						scim_id = scaleway_iam_scim.main.id
						token_id = scaleway_iam_scim_token.main.id
						depends_on = [scaleway_iam_scim_token.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "organization_id", "scaleway_iam_scim_token.main", "organization_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "scim_id", "scaleway_iam_scim_token.main", "scim_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "token_id", "scaleway_iam_scim_token.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "created_at", "scaleway_iam_scim_token.main", "created_at"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "expires_at", "scaleway_iam_scim_token.main", "expires_at"),
				),
			},
		},
	})
}

func TestAccDataSourceScimToken_WithoutScimID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			checkScimTokenDestroyed(tt),
			checkScimDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_scim" "main" {
						organization_id = "%[1]s"
					}

					resource "scaleway_iam_scim_token" "main" {
						scim_id = scaleway_iam_scim.main.id
						organization_id = "%[1]s"
					}

					data "scaleway_iam_scim_token" "main" {
						token_id = scaleway_iam_scim_token.main.id
						organization_id = "%[1]s"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "organization_id", "scaleway_iam_scim_token.main", "organization_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "scim_id", "scaleway_iam_scim_token.main", "scim_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "token_id", "scaleway_iam_scim_token.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "created_at", "scaleway_iam_scim_token.main", "created_at"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim_token.main", "expires_at", "scaleway_iam_scim_token.main", "expires_at"),
				),
			},
		},
	})
}

func TestAccDataSourceScimToken_InvalidToken(t *testing.T) {
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
					resource "scaleway_iam_scim" "main" {
						organization_id = "%s"
					}

					data "scaleway_iam_scim_token" "main" {
						scim_id = scaleway_iam_scim.main.id
						token_id = "00000000-0000-0000-0000-000000000000"
						organization_id = "%[1]s"
					}
				`, orgID),
				ExpectError: regexp.MustCompile("SCIM token.*not found"),
			},
		},
	})
}
