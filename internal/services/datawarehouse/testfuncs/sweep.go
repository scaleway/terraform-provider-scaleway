package testfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	datawarehouseSDK "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_datawarehouse_deployment", &resource.Sweeper{
		Name: "scaleway_datawarehouse_deployment",
		F:    testSweepDatawarehouseDeployment,
	})
}

func testSweepDatawarehouseDeployment(_ string) error {
	return acctest.SweepRegions([]scw.Region{scw.RegionFrPar}, func(scwClient *scw.Client, region scw.Region) error {
		datawarehouseAPI := datawarehouseSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying datawarehouse deployments in (%s)", region)

		listDeployments, err := datawarehouseAPI.ListDeployments(&datawarehouseSDK.ListDeploymentsRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing datawarehouse deployments in (%s) in sweeper: %w", region, err)
		}

		for _, deployment := range listDeployments.Deployments {
			_, err := datawarehouseAPI.DeleteDeployment(&datawarehouseSDK.DeleteDeploymentRequest{
				Region:       region,
				DeploymentID: deployment.ID,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error deleting datawarehouse deployment %s in (%s): %s", deployment.ID, region, err)

				continue
			}
		}

		return nil
	})
}
