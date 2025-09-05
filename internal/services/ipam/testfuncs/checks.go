package ipamtestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	ipam2 "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
)

func CheckIPDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_ipam_ip" {
					continue
				}

				ipamAPI, region, ID, err := ipam.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = ipamAPI.GetIP(&ipam2.GetIPRequest{
					IPID:   ID,
					Region: region,
				})
				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("IP (%s) still exists", rs.Primary.ID))
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
