package testfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	searchdbSDK "github.com/scaleway/scaleway-sdk-go/api/searchdb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_opensearch_deployment", &resource.Sweeper{
		Name: "scaleway_opensearch_deployment",
		F:    testSweepOpenSearchDeployment,
	})
}

func testSweepOpenSearchDeployment(_ string) error {
	return acctest.SweepRegions((&searchdbSDK.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		opensearchAPI := searchdbSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying opensearch deployments in (%s)", region)

		listDeployments, err := opensearchAPI.ListDeployments(&searchdbSDK.ListDeploymentsRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing opensearch deployments in (%s) in sweeper: %s", region, err)

			return nil
		}

		for _, deployment := range listDeployments.Deployments {
			_, err := opensearchAPI.DeleteDeployment(&searchdbSDK.DeleteDeploymentRequest{
				Region:       region,
				DeploymentID: deployment.ID,
			})
			if err != nil {
				logging.L.Warningf("error deleting opensearch deployment %s in (%s): %s", deployment.ID, region, err)
			}
		}

		return nil
	})
}
