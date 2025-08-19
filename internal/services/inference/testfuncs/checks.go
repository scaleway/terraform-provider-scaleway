package inferencetestfuncs

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	inferenceSDK "github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/inference"
)

var DestroyWaitTimeout = 3 * time.Minute

func IsDeploymentDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_inference_deployment" {
				continue
			}

			inferenceAPI, region, ID, err := inference.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			deployment, err := inferenceAPI.GetDeployment(&inferenceSDK.GetDeploymentRequest{
				Region:       region,
				DeploymentID: ID,
			})
			if err == nil {
				return fmt.Errorf("deployment %s (%s) still exists", deployment.Name, deployment.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}

func IsModelDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_inference_model" {
					continue
				}

				api, region, id, err := inference.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				model, err := api.GetModel(&inferenceSDK.GetModelRequest{
					Region:  region,
					ModelID: id,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("model %s (%s) still exists", model.Name, model.ID))
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
