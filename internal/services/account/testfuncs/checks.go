package accounttestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	account2 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

var DestroyWaitTimeout = 3 * time.Minute

func IsProjectDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()
		api := account.NewProjectAPI(tt.Meta)

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_account_project" {
					continue
				}

				_, err := api.GetProject(&account2.ProjectAPIGetProjectRequest{
					ProjectID: rs.Primary.ID,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("resource %s(%s) still exists", rs.Type, rs.Primary.ID))
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
