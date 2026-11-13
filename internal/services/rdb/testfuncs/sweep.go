package rdbtestfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_rdb_instance", &resource.Sweeper{
		Name: "scaleway_rdb_instance",
		F:    testSweepInstance,
	})
	resource.AddTestSweepers("scaleway_rdb_snapshot", &resource.Sweeper{
		Name: "scaleway_rdb_snapshot",
		F:    testSweepInstanceSnapshots,
	})
}

func testSweepInstance(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		rdbAPI := rdbSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the rdb instance in (%s)", region)

		listInstances, err := rdbAPI.ListInstances(&rdbSDK.ListInstancesRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing rdb instances in (%s) in sweeper: %s", region, err)

			return nil
		}

		for _, instance := range listInstances.Instances {
			_, err := rdbAPI.DeleteInstance(&rdbSDK.DeleteInstanceRequest{
				Region:     region,
				InstanceID: instance.ID,
			})
			if err != nil {
				logging.L.Warningf("error deleting rdb instance in sweeper: %s", err)
			}
		}

		return nil
	})
}

func testSweepInstanceSnapshots(_ string) error {
	return acctest.SweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		rdbAPI := rdbSDK.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the rdb snapshots in (%s)", region)

		listSnapshots, err := rdbAPI.ListSnapshots(&rdbSDK.ListSnapshotsRequest{
			Region: region,
		}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing rdb snapshots in (%s) in sweeper: %s", region, err)

			return nil
		}

		for _, snapshot := range listSnapshots.Snapshots {
			_, err := rdbAPI.DeleteSnapshot(&rdbSDK.DeleteSnapshotRequest{
				Region:     region,
				SnapshotID: snapshot.ID,
			})
			if err != nil {
				logging.L.Warningf("Failed to delete rdb snapshot: %s", err.Error())
			}
		}

		return nil
	})
}
