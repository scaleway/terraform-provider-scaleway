package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
)

func TestAccScalewayDataSourceIamUser_Basic(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_iam_user" "by_id" {
					  user_id = "af194b1f-55a7-43f2-b61c-22a0268559e3"
					  organization_id = "dd5b8103-52ef-40b6-b157-35a426650401"
					}

					data "scaleway_iam_user" "by_email" {
					  email = "developer-tools-team@scaleway.com"
					  organization_id = "dd5b8103-52ef-40b6-b157-35a426650401"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamUserExists(tt, "data.scaleway_iam_user.by_id"),
					testAccCheckScalewayIamUserExists(tt, "data.scaleway_iam_user.by_email"),

					resource.TestCheckResourceAttrSet("data.scaleway_iam_user.by_id", "user_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_iam_user.by_id", "email"),

					resource.TestCheckResourceAttrSet("data.scaleway_iam_user.by_email", "user_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_iam_user.by_email", "email"),
				),
			},
		},
	})
}

func testAccCheckScalewayIamUserExists(tt *TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		iamAPI := iamAPI(tt.Meta)

		_, err := iamAPI.GetUser(&iam.GetUserRequest{
			UserID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("could not find user: %w", err)
		}

		return nil
	}
}
