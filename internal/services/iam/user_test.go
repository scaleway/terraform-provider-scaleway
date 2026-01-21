package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
)

func TestAccUser_Member(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             iamchecks.CheckUserDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_user" "member_user" {
						email = "testiamusermember@scaleway.com"
						username = "testiamusermember"
						first_name = "Foo"
						last_name = "Bar"
						password = "Firstaccesspsw123"
						locale = "en_US"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.member_user"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.member_user", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "email", "testiamusermember@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.member_user", "username", "testiamusermember"),
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

func TestAccUser_PasswordWO(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             iamchecks.CheckUserDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_user" "password_wo_user" {
						email = "testiamuserpasswordwo@scaleway.com"
						username = "testiamuserpasswordwo"
						password_wo = "FirstWOPassword123"
						password_wo_version = 1
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.password_wo_user"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.password_wo_user", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_user.password_wo_user", "email", "testiamuserpasswordwo@scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_iam_user.password_wo_user", "username", "testiamuserpasswordwo"),
				),
			},
		},
	})
}
