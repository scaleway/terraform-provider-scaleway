package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
)

func TestAccDataSourceUser_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             iamchecks.CheckUserDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_user" "main" {
					  email = "testiamuserbasic@scaleway.com"
					  username = "testiamuserbasic"
					}

					data "scaleway_iam_user" "by_id" {
					  user_id = scaleway_iam_user.main.id
					}

					data "scaleway_iam_user" "by_email" {
					  email = scaleway_iam_user.main.email
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "data.scaleway_iam_user.by_id"),
					testAccCheckIamUserExists(tt, "data.scaleway_iam_user.by_email"),

					resource.TestCheckResourceAttrSet("data.scaleway_iam_user.by_id", "user_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_iam_user.by_id", "email"),

					resource.TestCheckResourceAttrSet("data.scaleway_iam_user.by_email", "user_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_iam_user.by_email", "email"),
				),
			},
		},
	})
}

func testAccCheckIamUserExists(tt *acctest.TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		iamAPI := iam.NewAPI(tt.Meta)

		_, err := iamAPI.GetUser(&iamSDK.GetUserRequest{
			UserID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("could not find user: %w", err)
		}

		return nil
	}
}
