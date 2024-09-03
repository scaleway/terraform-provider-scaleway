package inference

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func testSweepDeployment(_ string) error {
	return acctest.SweepRegions((&inference.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		inferenceAPI := inference.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the inference deployments in (%s)", region)
		listDeployments, err := inferenceAPI.ListDeployments(
			&inference.ListDeploymentsRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing deployment in (%s) in sweeper: %s", region, err)
		}

		for _, deployment := range listDeployments.Deployments {
			_, err := inferenceAPI.DeleteDeployment(&inference.DeleteDeploymentRequest{
				DeploymentID: deployment.ID,
				Region:      region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting deployment in sweeper: %s", err)
			}
		}

		return nil
	})
}