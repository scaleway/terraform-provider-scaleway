package container_test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	containerSDK "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func init() {
	resource.AddTestSweepers("scaleway_container_namespace", &resource.Sweeper{
		Name:         "scaleway_container_namespace",
		F:            testSweepNamespace,
		Dependencies: []string{"scaleway_container"},
	})
}

func testSweepNamespace(_ string) error {
	return acctest.SweepRegions([]scw.Region{scw.RegionFrPar}, func(scwClient *scw.Client, region scw.Region) error {
		containerAPI := containerSDK.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the container namespaces in (%s)", region)
		listNamespaces, err := containerAPI.ListNamespaces(
			&containerSDK.ListNamespacesRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing namespaces in (%s) in sweeper: %s", region, err)
		}

		for _, ns := range listNamespaces.Namespaces {
			_, err := containerAPI.DeleteNamespace(&containerSDK.DeleteNamespaceRequest{
				NamespaceID: ns.ID,
				Region:      region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting namespace in sweeper: %s", err)
			}
		}

		return nil
	})
}
