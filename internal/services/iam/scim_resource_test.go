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

func TestAccScimResource_Basic(t *testing.T) {
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
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScimResourceExists(tt, "scaleway_iam_scim.main"),
					resource.TestCheckResourceAttr("scaleway_iam_scim.main", "organization_id", orgID),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim.main", "created_at"),
				),
			},
			{
				ResourceName:      "scaleway_iam_scim.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScimResource_WithDefaultOrganizationID(t *testing.T) {
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
				Config: `
					resource "scaleway_iam_scim" "main" {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScimResourceExists(tt, "scaleway_iam_scim.main"),
					resource.TestCheckResourceAttr("scaleway_iam_scim.main", "organization_id", orgID),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim.main", "created_at"),
				),
			},
		},
	})
}

func checkScimDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_iam_scim" {
				continue
			}

			iamAPI := iam.NewAPI(tt.Meta)

			_, err := iamAPI.GetOrganizationScim(&iamSDK.GetOrganizationScimRequest{
				OrganizationID: rs.Primary.Attributes["organization_id"],
			})
			if err == nil {
				return fmt.Errorf("SCIM configuration (%s) still exists", rs.Primary.ID)
			}

			if httperrors.Is404(err) {
				continue
			}

			return err
		}

		return nil
	}
}

func testAccCheckScimResourceExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iamAPI := iam.NewAPI(tt.Meta)

		scim, err := iamAPI.GetOrganizationScim(&iamSDK.GetOrganizationScimRequest{
			OrganizationID: rs.Primary.Attributes["organization_id"],
		})
		if err != nil {
			return err
		}

		if scim.CreatedAt.String() != rs.Primary.Attributes["created_at"] {
			return fmt.Errorf("SCIM created_at mismatch: expected %s, got %s",
				rs.Primary.Attributes["created_at"], scim.CreatedAt)
		}

		if scim.ID != rs.Primary.Attributes["id"] {
			return fmt.Errorf("SCIM ID mismatch: expected %s, got %s",
				rs.Primary.Attributes["id"], scim.ID)
		}

		return nil
	}
}
