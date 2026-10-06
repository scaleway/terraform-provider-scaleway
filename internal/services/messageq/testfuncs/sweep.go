package testfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	messageqSDK "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_messageq_deployment", &resource.Sweeper{
		Name: "scaleway_messageq_deployment",
		F:    testSweepMessageQDeployment,
	})
}

func testSweepMessageQDeployment(_ string) error {
	return acctest.SweepRegions((&messageqSDK.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		messageqAPI := messageqSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying messageq deployments in (%s)", region)

		listDeployments, err := messageqAPI.ListDeployments(&messageqSDK.ListDeploymentsRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing messageq deployments in (%s) in sweeper: %s", region, err)

			return nil
		}

		for _, deployment := range listDeployments.Deployments {
			_, err := messageqAPI.DeleteDeployment(&messageqSDK.DeleteDeploymentRequest{
				Region:       region,
				DeploymentID: deployment.ID,
			})
			if err != nil {
				logging.L.Warningf("error deleting messageq deployment %s in (%s): %s", deployment.ID, region, err)
			}
		}

		return nil
	})
}
