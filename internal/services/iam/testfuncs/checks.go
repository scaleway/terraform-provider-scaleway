package iamtestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
)

var DestroyWaitTimeout = 3 * time.Minute

func CheckSSHKeyDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		api := iam.NewAPI(tt.Meta)
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_iam_ssh_key" {
					continue
				}

				_, err := api.GetSSHKey(&iamSDK.GetSSHKeyRequest{
					SSHKeyID: rs.Primary.ID,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("SSH key (%s) still exists", rs.Primary.ID))
				case httperrors.Is404(err):
					continue
				default:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}

func CheckSSHKeyExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iamAPI := iam.NewAPI(tt.Meta)

		_, err := iamAPI.GetSSHKey(&iamSDK.GetSSHKeyRequest{
			SSHKeyID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func CheckUserDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
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
