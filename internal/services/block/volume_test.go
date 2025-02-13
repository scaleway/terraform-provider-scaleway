package block_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestAccVolume_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      blocktestfuncs.IsVolumeDestroyed(tt),
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
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
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
		CheckDestroy:      blocktestfuncs.IsVolumeDestroyed(tt),
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
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "snapshot_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "size_in_gb", "scaleway_block_volume.base", "size_in_gb"),
				),
			},
		},
	})
}
