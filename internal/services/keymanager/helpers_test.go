package keymanager_test

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	secretSDK "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

const (
	destroyWaitTimeout = 3 * time.Minute
)

func testAccCheckSecretDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, destroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_secret" {
					continue
				}

				api, region, id, err := secret.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				sec, err := api.GetSecret(&secretSDK.GetSecretRequest{
					SecretID: id,
					Region:   region,
				})

				switch {
				case err == nil && sec != nil && sec.DeletionRequestedAt != nil:
					// Soft-deleted (scheduled for deletion), treat as destroyed for tests
					continue
				case httperrors.Is404(err):
					continue
				case err != nil:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}
