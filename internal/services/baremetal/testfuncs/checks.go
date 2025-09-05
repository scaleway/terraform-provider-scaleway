package baremetaltestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	baremetal2 "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/baremetal"
)

func CheckServerDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_baremetal_server" {
					continue
				}

				api, zonedID, err := baremetal.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetServer(&baremetal2.GetServerRequest{
					ServerID: zonedID.ID,
					Zone:     zonedID.Zone,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("server (%s) still exists", rs.Primary.ID))
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
