package block_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestAccDataSourceSnapshot_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			blocktestfuncs.IsSnapshotDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
					}

					resource scaleway_block_snapshot main {
						name = "test-ds-block-snapshot-basic-tf"
						volume_id = scaleway_block_volume.main.id
					}

					data scaleway_block_snapshot find_by_name {
						name = scaleway_block_snapshot.main.name
					}

					data scaleway_block_snapshot find_by_id {
						snapshot_id = scaleway_block_snapshot.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.main"),

					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "name", "data.scaleway_block_snapshot.find_by_name", "name",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "name", "data.scaleway_block_snapshot.find_by_id", "name",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "id", "data.scaleway_block_snapshot.find_by_name", "id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "id", "data.scaleway_block_snapshot.find_by_id", "id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "zone", "data.scaleway_block_snapshot.find_by_name", "zone",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "zone", "data.scaleway_block_snapshot.find_by_id", "zone",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "volume_id", "data.scaleway_block_snapshot.find_by_name", "volume_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "volume_id", "data.scaleway_block_snapshot.find_by_id", "volume_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "tags", "data.scaleway_block_snapshot.find_by_name", "tags",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "tags", "data.scaleway_block_snapshot.find_by_id", "tags",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "project_id", "data.scaleway_block_snapshot.find_by_name", "project_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "project_id", "data.scaleway_block_snapshot.find_by_id", "project_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "import", "data.scaleway_block_snapshot.find_by_name", "import",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "import", "data.scaleway_block_snapshot.find_by_id", "import",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "export", "data.scaleway_block_snapshot.find_by_name", "export",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "export", "data.scaleway_block_snapshot.find_by_id", "export",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "id", "data.scaleway_block_snapshot.find_by_name", "snapshot_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "id", "data.scaleway_block_snapshot.find_by_id", "snapshot_id",
					),
				),
			},
		},
	})
}
