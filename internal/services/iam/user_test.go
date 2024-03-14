package iam_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
)

func init() {
	resource.AddTestSweepers("scaleway_iam_user", &resource.Sweeper{
		Name: "scaleway_iam_user",
		F:    testSweepIamUser,
	})
}

func testSweepIamUser(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		api := iamSDK.NewAPI(scwClient)

		orgID, exists := scwClient.GetDefaultOrganizationID()
		if !exists {
			return errors.New("missing organizationID")
		}

		listUsers, err := api.ListUsers(&iamSDK.ListUsersRequest{
			OrganizationID: &orgID,
		})
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}
		for _, user := range listUsers.Users {
			if !acctest.IsTestResource(user.Email) {
				continue
			}
			err = api.DeleteUser(&iamSDK.DeleteUserRequest{
				UserID: user.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete user: %w", err)
			}
		}
		return nil
	})
}

func TestAccIamUser_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckIamUserDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_user" "user_basic" {
							email = "foo@scaleway.com"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists(tt, "scaleway_iam_user.user_basic"),
					acctest.CheckResourceAttrUUID("scaleway_iam_user.user_basic", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_user.user_basic", "email", "foo@scaleway.com"),
				),
			},
		},
	})
}

func testAccCheckIamUserDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
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
