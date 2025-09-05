package vpcgwtestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	vpcgwSDK "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw"
)

func IsGatewayNetworkDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc_gateway_network" {
					continue
				}
				api, zone, id, err := vpcgw.NewAPIWithZoneAndIDv2(tt.Meta, rs.Primary.ID)
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

		return retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc_public_gateway" {
					continue
				}
				api, zone, id, err := vpcgw.NewAPIWithZoneAndIDv2(tt.Meta, rs.Primary.ID)
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

func IsDHCPDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc_public_gateway_dhcp" {
					continue
				}
				api, zone, id, err := vpcgw.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID) // v1 API helper
				if err != nil {
					return retry.NonRetryableError(err)
				}
				_, err = api.GetDHCP(&vpcgwSDK.GetDHCPRequest{
					DHCPID: id,
					Zone:   zone,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("VPC public gateway DHCP config (%s) still exists", rs.Primary.ID))
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

		return retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc_public_gateway_ip" {
					continue
				}
				api, zone, id, err := vpcgw.NewAPIWithZoneAndIDv2(tt.Meta, rs.Primary.ID)
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
