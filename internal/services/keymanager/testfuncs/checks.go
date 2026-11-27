package keymanagertestfuncs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/keymanager"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func IsKeyManagerKeyDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_key_manager_key" {
				continue
			}

			client, region, keyID, err := keymanager.NewKeyManagerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			key, err := transport.RetryOnStatusValue(tt.T.Context(), transport.IAMPropagation, func(ctx context.Context) (*key_manager.Key, error) {
				return client.GetKey(&key_manager.GetKeyRequest{
					Region: region,
					KeyID:  keyID,
				})
			})
			if err == nil {
				if key.DeletionRequestedAt != nil {
					continue
				}

				return fmt.Errorf("key manager key (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
