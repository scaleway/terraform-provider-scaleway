package vpcgwtestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw"
)

var DestroyWaitTimeout = 3 * time.Minute

func IsGatewayNetworkDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc_gateway_network" {
					continue
				}

				api, zone, id, err := vpcgw.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetGatewayNetwork(&v2.GetGatewayNetworkRequest{
					GatewayNetworkID: id,
					Zone:             zone,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("VPC gateway network (%s) still exists", rs.Primary.ID))
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

func IsGatewayDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc_public_gateway" {
					continue
				}

				api, zone, id, err := vpcgw.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetGateway(&v2.GetGatewayRequest{
					GatewayID: id,
					Zone:      zone,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("VPC public gateway (%s) still exists", rs.Primary.ID))
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

func IsIPDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc_public_gateway_ip" {
					continue
				}

				api, zone, id, err := vpcgw.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetIP(&v2.GetIPRequest{
					IPID: id,
					Zone: zone,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("VPC public gateway ip (%s) still exists", rs.Primary.ID))
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
