package block_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	blockSDK "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block"
)

func init() {
	resource.AddTestSweepers("scaleway_block_volume", &resource.Sweeper{
		Name: "scaleway_block_volume",
		F:    testSweepBlockVolume,
	})
}

func testSweepBlockVolume(_ string) error {
	return acctest.SweepZones((&blockSDK.API{}).Zones(), func(scwClient *scw.Client, zone scw.Zone) error {
		blockAPI := blockSDK.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the block volumes in (%s)", zone)
		listVolumes, err := blockAPI.ListVolumes(
			&blockSDK.ListVolumesRequest{
				Zone: zone,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing volume in (%s) in sweeper: %s", zone, err)
		}

		for _, volume := range listVolumes.Volumes {
			err := blockAPI.DeleteVolume(&blockSDK.DeleteVolumeRequest{
				VolumeID: volume.ID,
				Zone:     zone,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting volume in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccVolume_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						name = "test-block-volume-basic"
						iops = 5000
						size_in_gb = 20
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isVolumePresent(tt, "scaleway_block_volume.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "name", "test-block-volume-basic"),
					resource.TestCheckResourceAttr("scaleway_block_volume.main", "size_in_gb", "20"),
				),
			},
		},
	})
}

func TestAccVolume_FromSnapshot(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isVolumeDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume base {
						name = "test-block-volume-from-snapshot-base"
						iops = 5000
						size_in_gb = 20
					}

					resource scaleway_block_snapshot main {
						name = "test-block-volume-from-snapshot"
						volume_id = scaleway_block_volume.base.id
					}

					resource scaleway_block_volume main {
						name = "test-block-volume-from-snapshot"
						iops = 5000
						snapshot_id = scaleway_block_snapshot.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isVolumePresent(tt, "scaleway_block_volume.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "snapshot_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "size_in_gb", "scaleway_block_volume.base", "size_in_gb"),
				),
			},
		},
	})
}

func isVolumePresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, zone, id, err := block.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetVolume(&blockSDK.GetVolumeRequest{
			VolumeID: id,
			Zone:     zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isVolumeDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_block_volume" {
				continue
			}

			api, zone, id, err := block.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteVolume(&blockSDK.DeleteVolumeRequest{
				VolumeID: id,
				Zone:     zone,
			})

			if err == nil {
				return fmt.Errorf("block volume (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) && !httperrors.Is410(err) {
				return err
			}
		}

		return nil
	}
}
