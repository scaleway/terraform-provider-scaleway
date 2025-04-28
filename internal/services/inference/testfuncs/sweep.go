package inferencetestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/scaleway-sdk-go/api/inference/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_inference_deployment", &resource.Sweeper{
		Name:         "scaleway_inference_deployment",
		Dependencies: nil,
		F:            testSweepDeployment,
	})
	resource.AddTestSweepers("scaleway_inference_custom_model", &resource.Sweeper{
		Name:         "scaleway_inference_custom_model",
		Dependencies: nil,
		F:            testSweepCustomModel,
	})
}

func testSweepDeployment(_ string) error {
	return acctest.SweepRegions((&inference.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		inferenceAPI := inference.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the inference deployments in (%s)", region)

		listDeployments, err := inferenceAPI.ListDeployments(
			&inference.ListDeploymentsRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing deployment in (%s) in sweeper: %w", region, err)
		}

		for _, deployment := range listDeployments.Deployments {
			_, err := inferenceAPI.DeleteDeployment(&inference.DeleteDeploymentRequest{
				DeploymentID: deployment.ID,
				Region:       region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting deployment in sweeper: %w", err)
			}
		}

		return nil
	})
}

func testSweepCustomModel(_ string) error {
	return acctest.SweepRegions((&inference.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		inferenceAPI := inference.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the inference models in (%s)", region)

		listModels, err := inferenceAPI.ListModels(&inference.ListModelsRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing models in (%s) in sweeper: %w", region, err)
		}

		for _, model := range listModels.Models {
			err := inferenceAPI.DeleteModel(&inference.DeleteModelRequest{
				Region:  model.Region,
				ModelID: model.ID,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting model in sweeper: %w", err)
			}
		}

		return nil
	})
}
