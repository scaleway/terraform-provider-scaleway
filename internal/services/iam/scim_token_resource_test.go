package iam_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
)

func TestAccScimTokenResource_Basic(t *testing.T) {
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
						organization_id = "%s"
					}
				`, orgID, orgID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScimTokenResourceExists(tt, "scaleway_iam_scim_token.main"),
					resource.TestCheckResourceAttr("scaleway_iam_scim_token.main", "organization_id", orgID),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim_token.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim_token.main", "bearer_token"),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim_token.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim_token.main", "expires_at"),
					resource.TestCheckResourceAttrPair("scaleway_iam_scim_token.main", "scim_id", "scaleway_iam_scim.main", "id"),
				),
			},
			{
				ResourceName:            "scaleway_iam_scim_token.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bearer_token"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					tokenID := state.RootModule().Resources["scaleway_iam_scim_token.main"].Primary.ID

					return tokenID, nil
				},
			},
			{
				ResourceName:            "scaleway_iam_scim_token.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bearer_token"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					orgID := state.RootModule().Resources["scaleway_iam_scim_token.main"].Primary.Attributes["organization_id"]
					tokenID := state.RootModule().Resources["scaleway_iam_scim_token.main"].Primary.ID

					return fmt.Sprintf("%s/%s", orgID, tokenID), nil
				},
			},
		},
	})
}

func TestAccScimTokenResource_WithDefaultOrganizationID(t *testing.T) {
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
				Config: `
					resource "scaleway_iam_scim" "main" {
					}

					resource "scaleway_iam_scim_token" "main" {
						scim_id = scaleway_iam_scim.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScimTokenResourceExists(tt, "scaleway_iam_scim_token.main"),
					resource.TestCheckResourceAttr("scaleway_iam_scim_token.main", "organization_id", orgID),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim_token.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim_token.main", "bearer_token"),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim_token.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_iam_scim_token.main", "expires_at"),
					resource.TestCheckResourceAttrPair("scaleway_iam_scim_token.main", "scim_id", "scaleway_iam_scim.main", "id"),
				),
			},
		},
	})
}

func checkScimTokenDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_iam_scim_token" {
				continue
			}

			iamAPI := iam.NewAPI(tt.Meta)

			_, err := iamAPI.GetOrganizationScim(&iamSDK.GetOrganizationScimRequest{
				OrganizationID: rs.Primary.Attributes["organization_id"],
			})
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return fmt.Errorf("failed to get SCIM configuration: %w", err)
			}

			listResp, err := iamAPI.ListScimTokens(&iamSDK.ListScimTokensRequest{
				ScimID: rs.Primary.Attributes["scim_id"],
			})
			if err != nil {
				return fmt.Errorf("failed to list SCIM tokens: %w", err)
			}

			deletedTokenID := rs.Primary.ID
			for _, token := range listResp.ScimTokens {
				if token.ID == deletedTokenID {
					return fmt.Errorf("SCIM token %s still exists after deletion", deletedTokenID)
				}
			}
		}

		return nil
	}
}

func testAccCheckScimTokenResourceExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iamAPI := iam.NewAPI(tt.Meta)

		listResp, err := iamAPI.ListScimTokens(&iamSDK.ListScimTokensRequest{
			ScimID: rs.Primary.Attributes["scim_id"],
		})
		if err != nil {
			return fmt.Errorf("failed to list SCIM tokens: %w", err)
		}

		tokenID := rs.Primary.ID
		found := false

		for _, token := range listResp.ScimTokens {
			if token.ID == tokenID {
				found = true

				break
			}
		}

		if !found {
			return fmt.Errorf("SCIM token %s not found in the list", tokenID)
		}

		if rs.Primary.ID == "" {
			return errors.New("SCIM token ID is not set")
		}

		if rs.Primary.Attributes["bearer_token"] == "" {
			return errors.New("SCIM token bearer_token is not set")
		}

		return nil
	}
}
