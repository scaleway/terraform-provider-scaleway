package testfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	kafkaSDK "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/kafka"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_kafka_cluster", &resource.Sweeper{
		Name: "scaleway_kafka_cluster",
		F:    testSweepCluster,
	})
}

func testSweepCluster(region string) error {
	return acctest.SweepRegions([]scw.Region{scw.Region(region)}, func(scwClient *scw.Client, region scw.Region) error {
		api := kafka.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting kafka clusters in (%s)", region)

		listClusters, err := api.ListClusters(&kafkaSDK.ListClustersRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing kafka clusters in sweeper: %w", err)
		}

		for _, cluster := range listClusters.Clusters {
			_, err := api.DeleteCluster(&kafkaSDK.DeleteClusterRequest{
				Region:    region,
				ClusterID: cluster.ID,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting kafka cluster in sweeper: %w", err)
			}
		}

		return nil
	})
}
