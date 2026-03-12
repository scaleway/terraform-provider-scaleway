package vpctestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	vpc2 "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc"
)

var DestroyWaitTimeout = 3 * time.Minute

func CheckPrivateNetworkDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc_private_network" {
					continue
				}

				api, region, id, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetPrivateNetwork(&vpc2.GetPrivateNetworkRequest{
					Region:           region,
					PrivateNetworkID: id,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("VPC private network (%s) still exists", rs.Primary.ID))
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

func IsPrivateNetworkPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcAPI, region, ID, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcAPI.GetPrivateNetwork(&vpc2.GetPrivateNetworkRequest{
			PrivateNetworkID: ID,
			Region:           region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func IsVPCPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		vpcAPI, region, ID, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = vpcAPI.GetVPC(&vpc2.GetVPCRequest{
			VpcID:  ID,
			Region: region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func CheckVPCDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_vpc" {
					continue
				}

				vpcAPI, region, id, err := vpc.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = vpcAPI.GetVPC(&vpc2.GetVPCRequest{
					Region: region,
					VpcID:  id,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("VPC (%s) still exists", rs.Primary.ID))
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
