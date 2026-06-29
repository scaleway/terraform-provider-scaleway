package block_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestAccDataSourceVolume_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			blocktestfuncs.IsVolumeDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
  						name = "test-ds-block-volume-basic"
					}

					data scaleway_block_volume find_by_name {
						name = scaleway_block_volume.main.name
					}

					data scaleway_block_volume find_by_id {
						volume_id = scaleway_block_volume.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.main"),

					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "name", "data.scaleway_block_volume.find_by_name", "name",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "name", "data.scaleway_block_volume.find_by_id", "name",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "id", "data.scaleway_block_volume.find_by_name", "id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "id", "data.scaleway_block_volume.find_by_id", "id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "project_id", "data.scaleway_block_volume.find_by_name", "project_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "project_id", "data.scaleway_block_volume.find_by_id", "project_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "tags", "data.scaleway_block_volume.find_by_name", "tags",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "tags", "data.scaleway_block_volume.find_by_id", "tags",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "size_in_gb", "data.scaleway_block_volume.find_by_name", "size_in_gb",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "size_in_gb", "data.scaleway_block_volume.find_by_id", "size_in_gb",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "zone", "data.scaleway_block_volume.find_by_name", "zone",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "zone", "data.scaleway_block_volume.find_by_id", "zone",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "iops", "data.scaleway_block_volume.find_by_name", "iops",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "iops", "data.scaleway_block_volume.find_by_id", "iops",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "snapshot_id", "data.scaleway_block_volume.find_by_name", "snapshot_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "snapshot_id", "data.scaleway_block_volume.find_by_id", "snapshot_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "id", "data.scaleway_block_volume.find_by_name", "volume_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_volume.main", "id", "data.scaleway_block_volume.find_by_id", "volume_id",
					),
				),
			},
		},
	})
}
