package annotationstestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
)

var DestroyWaitTimeout = 3 * time.Minute

func IsAnnotationsKeyDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_annotations_key" {
					continue
				}

				api := annotationsSDK.NewAPI(tt.Meta.ScwClient())
				_, err := api.GetKey(&annotationsSDK.GetKeyRequest{
					KeyID: rs.Primary.ID,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("annotation key (%s) still exists", rs.Primary.ID))
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

func IsAnnotationsValueDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_annotations_value" {
					continue
				}

				api := annotationsSDK.NewAPI(tt.Meta.ScwClient())
				_, err := api.GetValue(&annotationsSDK.GetValueRequest{
					ValueID: rs.Primary.ID,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("annotation value (%s) still exists", rs.Primary.ID))
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

func IsAnnotationsBindingDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_annotations_binding" {
					continue
				}

				api := annotationsSDK.NewAPI(tt.Meta.ScwClient())

				bindings, err := api.ListBindings(&annotationsSDK.ListBindingsRequest{})
				if err != nil {
					return retry.NonRetryableError(err)
				}

				for _, binding := range bindings.Bindings {
					if binding.ID == rs.Primary.ID {
						return retry.RetryableError(fmt.Errorf("annotation binding (%s) still exists", rs.Primary.ID))
					}
				}
			}

			return nil
		})
	}
}
