package iam_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceScim_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             checkScimDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_scim" "main" {
						organization_id = "%s"
					}

					data "scaleway_iam_scim" "main" {
						organization_id = scaleway_iam_scim.main.organization_id
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim.main", "organization_id", "scaleway_iam_scim.main", "organization_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim.main", "id", "scaleway_iam_scim.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_scim.main", "created_at", "scaleway_iam_scim.main", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceScim_WithDefaultOrganizationID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	_, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}
	{
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: tt.ProviderFactories,
			CheckDestroy:             checkScimDestroyed(tt),
			Steps: []resource.TestStep{
				{
					Config: `
					resource "scaleway_iam_scim" "main" {
					}

					data "scaleway_iam_scim" "main" {
						depends_on = [scaleway_iam_scim.main]
					}
				`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("data.scaleway_iam_scim.main", "organization_id", "scaleway_iam_scim.main", "organization_id"),
						resource.TestCheckResourceAttrPair("data.scaleway_iam_scim.main", "id", "scaleway_iam_scim.main", "id"),
						resource.TestCheckResourceAttrPair("data.scaleway_iam_scim.main", "created_at", "scaleway_iam_scim.main", "created_at"),
					),
				},
			},
		})
	}
}

func TestAccDataSourceScim_InvalidDeactivated(t *testing.T) {
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
					data "scaleway_iam_scim" "main" {
						organization_id = "%s"
					}
				`, orgID),
				ExpectError: regexp.MustCompile("not found"),
			},
		},
	})
}
