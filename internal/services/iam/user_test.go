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

func TestAccUser_Member(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isUserDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_user" "member_user" {
						email = "foo@scaleway.com"
						username = "foo"
						first_name = "Foo"
						last_name = "Bar"
						password = "Firstaccesspsw123"
						locale = "en_US"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.member_user"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.member_user", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "email", "foo@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "username", "foo"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "first_name", "Foo"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "last_name", "Bar"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "password", "Firstaccesspsw123"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "phone_number", ""),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "locale", "en_US"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "type", "member"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "mfa", "false"),
				),
			},
			// Add tag and update email, username, last name, phone number and locale.
			{
				Config: `
					resource "scaleway_iam_user" "member_user" {
						email = "foobar@scaleway.com"
						username = "foobar"
						last_name = "Baz"
						phone_number = "+33112345678"
						locale = "fr_FR"
						tags = ["tf_tests"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.member_user"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.member_user", "id"),
					// Let's check that a field is set to empty when not defined in the configuration.
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "first_name", ""),
					// Let's update some fields and test if they actually get updated.
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "email", "foobar@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "username", "foobar"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "last_name", "Baz"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "phone_number", "+33112345678"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "locale", "fr_FR"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "tags.0", "tf_tests"),
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
