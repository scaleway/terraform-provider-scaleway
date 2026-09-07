package blocktestfuncs

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_block_snapshot", &resource.Sweeper{
		Name: "scaleway_block_snapshot",
		F:    testSweepSnapshot,
	})
	resource.AddTestSweepers("scaleway_block_volume", &resource.Sweeper{
		Name: "scaleway_block_volume",
		F:    testSweepBlockVolume,
	})
}

func testSweepBlockVolume(_ string) error {
	return acctest.SweepZones((&block.API{}).Zones(), func(scwClient *scw.Client, zone scw.Zone) error {
		blockAPI := block.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the block volumes in (%s)", zone)

		listVolumes, err := blockAPI.ListVolumes(
			&block.ListVolumesRequest{
				Zone: zone,
			}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing volume in (%s) in sweeper: %s", zone, err)

			return nil
		}

		for _, volume := range listVolumes.Volumes {
			err := blockAPI.DeleteVolume(&block.DeleteVolumeRequest{
				VolumeID: volume.ID,
				Zone:     zone,
			})
			if err != nil {
				logging.L.Warningf("error deleting volume in sweeper: %s", err)
			}
		}

		return nil
	})
}

func testSweepSnapshot(_ string) error {
	return acctest.SweepZones((&block.API{}).Zones(), func(scwClient *scw.Client, zone scw.Zone) error {
		blockAPI := block.NewAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the block snapshots in (%s)", zone)

		listSnapshots, err := blockAPI.ListSnapshots(
			&block.ListSnapshotsRequest{
				Zone: zone,
			}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing snapshot in (%s) in sweeper: %s", zone, err)

			return nil
		}

		for _, snapshot := range listSnapshots.Snapshots {
			err := blockAPI.DeleteSnapshot(&block.DeleteSnapshotRequest{
				SnapshotID: snapshot.ID,
				Zone:       zone,
			})
			if err != nil {
				logging.L.Warningf("error deleting snapshot in sweeper: %s", err)
			}
		}

		return nil
	})
}
