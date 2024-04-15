package documentdbtestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	documentdbSDK "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_documentdb_instance", &resource.Sweeper{
		Name: "scaleway_documentdb_instance",
		F:    testSweepInstance,
	})
}

func testSweepInstance(_ string) error {
	return acctest.SweepRegions((&documentdbSDK.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		api := documentdbSDK.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the documentdb instances in (%s)", region)
		listInstances, err := api.ListInstances(
			&documentdbSDK.ListInstancesRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing instance in (%s) in sweeper: %s", region, err)
		}

		for _, instance := range listInstances.Instances {
			_, err := api.DeleteInstance(&documentdbSDK.DeleteInstanceRequest{
				InstanceID: instance.ID,
				Region:     region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting instance in sweeper: %s", err)
			}
		}

		return nil
	})
}
