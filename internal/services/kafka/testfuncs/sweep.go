package testfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	kafkaSDK "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_kafka_cluster", &resource.Sweeper{
		Name: "scaleway_kafka_cluster",
		F:    testSweepCluster,
	})
}

func testSweepCluster(_ string) error {
	return acctest.SweepRegions((&kafkaSDK.API{}).Regions(), func(scwClient *scw.Client, region scw.Region) error {
		api := kafkaSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting kafka clusters in (%s)", region)

		listClusters, err := api.ListClusters(&kafkaSDK.ListClustersRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing kafka clusters in sweeper: %s", err)

			return nil
		}

		for _, cluster := range listClusters.Clusters {
			if !acctest.IsTestResource(cluster.Name) {
				continue
			}

			_, err := api.DeleteCluster(&kafkaSDK.DeleteClusterRequest{
				Region:    region,
				ClusterID: cluster.ID,
			})
			if err != nil {
				logging.L.Warningf("error deleting kafka cluster in sweeper: %s", err)
			}
		}

		return nil
	})
}
