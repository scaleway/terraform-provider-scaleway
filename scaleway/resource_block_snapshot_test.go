package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/errs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func init() {
	resource.AddTestSweepers("scaleway_block_snapshot", &resource.Sweeper{
		Name: "scaleway_block_snapshot",
		F:    testSweepBlockSnapshot,
	})
}

func testSweepBlockSnapshot(_ string) error {
	return sweepZones((&block.API{}).Zones(), func(scwClient *scw.Client, zone scw.Zone) error {
		blockAPI := block.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the block snapshots in (%s)", zone)
		listSnapshots, err := blockAPI.ListSnapshots(
			&block.ListSnapshotsRequest{
				Zone: zone,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing snapshot in (%s) in sweeper: %s", zone, err)
		}

		for _, snapshot := range listSnapshots.Snapshots {
			err := blockAPI.DeleteSnapshot(&block.DeleteSnapshotRequest{
				SnapshotID: snapshot.ID,
				Zone:       zone,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting snapshot in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayBlockSnapshot_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayBlockSnapshotDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
					}

					resource scaleway_block_snapshot main {
						name = "test-block-snapshot-basic"
						volume_id = scaleway_block_volume.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBlockSnapshotExists(tt, "scaleway_block_snapshot.main"),
					testCheckResourceAttrUUID("scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_snapshot.main", "name", "test-block-snapshot-basic"),
				),
			},
		},
	})
}

func testAccCheckScalewayBlockSnapshotExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := blockAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetSnapshot(&block.GetSnapshotRequest{
			SnapshotID: id,
			Zone:       zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayBlockSnapshotDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_block_snapshot" {
				continue
			}

			api, zone, id, err := blockAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteSnapshot(&block.DeleteSnapshotRequest{
				SnapshotID: id,
				Zone:       zone,
			})

			if err == nil {
				return fmt.Errorf("block snapshot (%s) still exists", rs.Primary.ID)
			}

			if !errs.Is404Error(err) {
				return err
			}
		}

		return nil
	}
}
