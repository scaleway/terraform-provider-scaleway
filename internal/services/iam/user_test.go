package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
)

func TestAccUser_Guest(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isUserDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_user" "guest_user" {
							email = "foo@scaleway.com"
							tags = ["tf_tests", "tests"]
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.guest_user"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.guest_user", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "email", "foo@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "tags.0", "tf_tests"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "tags.1", "tests"),
					// The username is the same as the email for Guest users.
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "username", "foo@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "type", "guest"),
				),
			},
			// Update tags
			{
				Config: `
						resource "scaleway_iam_user" "guest_user" {
							email = "foo@scaleway.com"
							tags = ["tf_tests"]
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.guest_user"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.guest_user", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "email", "foo@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "tags.0", "tf_tests"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "type", "guest"),
				),
			},
			// Remove tags
			{
				Config: `
						resource "scaleway_iam_user" "guest_user" {
							email = "foo@scaleway.com"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.guest_user"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.guest_user", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "email", "foo@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_iam_user.guest_user", "type", "guest"),
				),
			},
		},
	})
}

func TestAccUser_Member(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isUserDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_user" "member_user" {
							email = "foo@scaleway.com"
							username = "foo"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.member_user"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.member_user", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "email", "foo@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "username", "foo"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "type", "member"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "mfa", "false"),
				),
			},
			// Add tag
			{
				Config: `
						resource "scaleway_iam_user" "member_user" {
							email = "foo@scaleway.com"
							tags = ["tf_tests"]
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.member_user"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.member_user", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "email", "foo@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "tags.0", "tf_tests"),
					// Let's check the username isn't removed even when not defined in the configuration.
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "username", "foo"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "type", "member"),
				),
			},
		},
	})
}

func isUserDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_iam_user" {
				continue
			}

			iamAPI := iam.NewAPI(tt.Meta)

			_, err := iamAPI.GetUser(&iamSDK.GetUserRequest{
				UserID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("resource %s(%s) still exist", rs.Type, rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
