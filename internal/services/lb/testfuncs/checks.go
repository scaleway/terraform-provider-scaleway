package lbtestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	lb2 "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

var DestroyWaitTimeout = 3 * time.Minute

func IsIPDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_lb_ip" {
					continue
				}

				lbAPI, zone, ID, err := lb.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				if lbID, ok := rs.Primary.Attributes["lb_id"]; ok && lbID != "" {
					retryInterval := lb.DefaultWaitLBRetryInterval
					if transport.DefaultWaitRetryInterval != nil {
						retryInterval = *transport.DefaultWaitRetryInterval
					}

					_, waitErr := lbAPI.WaitForLbInstances(&lb2.ZonedAPIWaitForLBInstancesRequest{
						Zone:          zone,
						LBID:          lbID,
						Timeout:       scw.TimeDurationPtr(instance.DefaultInstanceServerWaitTimeout),
						RetryInterval: &retryInterval,
					}, scw.WithContext(ctx))

					// Unexpected api error we return it
					if waitErr != nil && !httperrors.Is404(waitErr) {
						return retry.NonRetryableError(waitErr)
					}
				}

				_, err = lbAPI.GetIP(&lb2.ZonedAPIGetIPRequest{
					Zone: zone,
					IPID: ID,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("IP (%s) still exists", rs.Primary.ID))
				case httperrors.Is403(err):
					return retry.RetryableError(err)
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
