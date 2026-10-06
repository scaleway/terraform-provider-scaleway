package mongodbtestfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_mongodb_instance", &resource.Sweeper{
		Name: "scaleway_mongodb_instance",
		F:    testSweepMongodbInstance,
	})
	resource.AddTestSweepers("scaleway_mongodb_instance_snapshot", &resource.Sweeper{
		Name: "scaleway_mongodb_instance_snapshot",
		F:    testSweepMongodbInstanceSnapshot,
	})
}

func testSweepMongodbInstance(_ string) error {
	return acctest.SweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		mongodbAPI := mongodb.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the mongodb instance in (%s)", zone)

		extractRegion, err := zone.Region()
		if err != nil {
			logging.L.Warningf("error extract region in (%s) in sweeper: %s", zone, err)

			return nil
		}

		listInstance, err := mongodbAPI.ListInstances(&mongodb.ListInstancesRequest{
			Region: extractRegion,
		})
		if err != nil {
			logging.L.Warningf("error listing mongodb instance in (%s) in sweeper: %s", zone, err)

			return nil
		}

		for _, instance := range listInstance.Instances {
			_, err := mongodbAPI.DeleteInstance(&mongodb.DeleteInstanceRequest{
				Region:     extractRegion,
				InstanceID: instance.ID,
			})
			if err != nil {
				logging.L.Warningf("error deleting mongodb instance %s in sweeper: %s", instance.ID, err)
			}
		}

		return nil
	})
}

func testSweepMongodbInstanceSnapshot(_ string) error {
	return acctest.SweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		mongodbAPI := mongodb.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the mongodb instance snapshot in (%s)", zone)

		extractRegion, err := zone.Region()
		if err != nil {
			logging.L.Warningf("error extract region in (%s) in sweeper: %s", zone, err)

			return nil
		}

		listSnapshot, err := mongodbAPI.ListSnapshots(&mongodb.ListSnapshotsRequest{
			Region: extractRegion,
		})
		if err != nil {
			logging.L.Warningf("error listing mongodb instance snapshot in (%s) in sweeper: %s", zone, err)

			return nil
		}

		for _, snapshot := range listSnapshot.Snapshots {
			_, err := mongodbAPI.DeleteSnapshot(&mongodb.DeleteSnapshotRequest{
				Region:     extractRegion,
				SnapshotID: snapshot.ID,
			})
			if err != nil {
				logging.L.Warningf("error deleting mongodb instance snapshot %s in sweeper: %s", snapshot.ID, err)
			}
		}

		return nil
	})
}
