package block_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestAccDataSourceSnapshots_Basic(t *testing.T) {
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
						name = "test-ds-block-snapshots-basic-tf"
						volume_id = scaleway_block_volume.main.id
						tags = ["test-ds-block-snapshots-basic-tf", "first"]
					}

					resource scaleway_block_snapshot second {
						name = "test-ds-block-snapshots-other-tf"
						volume_id = scaleway_block_volume.main.id
						tags = ["test-ds-block-snapshots-basic-tf", "second"]

						# The Block API rejects concurrent snapshot creations on the same volume
						depends_on = [scaleway_block_snapshot.main]
					}

					data scaleway_block_snapshots by_name {
						name = scaleway_block_snapshot.main.name
						depends_on = [scaleway_block_snapshot.main, scaleway_block_snapshot.second]
					}

					data scaleway_block_snapshots by_tags {
						tags = ["test-ds-block-snapshots-basic-tf"]
						depends_on = [scaleway_block_snapshot.main, scaleway_block_snapshot.second]
					}

					data scaleway_block_snapshots by_volume_id {
						volume_id = scaleway_block_volume.main.id
						depends_on = [scaleway_block_snapshot.main, scaleway_block_snapshot.second]
					}

					data scaleway_block_snapshots by_tags_ordered_desc {
						tags = ["test-ds-block-snapshots-basic-tf"]
						order_by = "created_at_desc"
						depends_on = [scaleway_block_snapshot.main, scaleway_block_snapshot.second]
					}

					data scaleway_block_snapshots by_name_empty {
						name = "test-ds-block-snapshots-nonexistent-tf"
						depends_on = [scaleway_block_snapshot.main, scaleway_block_snapshot.second]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.main"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.second"),

					resource.TestCheckResourceAttr("data.scaleway_block_snapshots.by_name", "snapshots.#", "1"),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "id", "data.scaleway_block_snapshots.by_name", "snapshots.0.id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "name", "data.scaleway_block_snapshots.by_name", "snapshots.0.name",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "zone", "data.scaleway_block_snapshots.by_name", "snapshots.0.zone",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "volume_id", "data.scaleway_block_snapshots.by_name", "snapshots.0.volume_id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "project_id", "data.scaleway_block_snapshots.by_name", "snapshots.0.project_id",
					),
					resource.TestCheckResourceAttr("data.scaleway_block_snapshots.by_name", "snapshots.0.tags.#", "2"),
					resource.TestMatchResourceAttr("data.scaleway_block_snapshots.by_name", "snapshots.0.srn", regexp.MustCompile(`^srn://block\..+/zones/.+/snapshots/.+$`)),

					resource.TestCheckResourceAttr("data.scaleway_block_snapshots.by_tags", "snapshots.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_block_snapshots.by_tags", "snapshots.*.id", "scaleway_block_snapshot.main", "id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_block_snapshots.by_tags", "snapshots.*.id", "scaleway_block_snapshot.second", "id",
					),

					resource.TestCheckResourceAttr("data.scaleway_block_snapshots.by_volume_id", "snapshots.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_block_snapshots.by_volume_id", "snapshots.*.id", "scaleway_block_snapshot.main", "id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"data.scaleway_block_snapshots.by_volume_id", "snapshots.*.id", "scaleway_block_snapshot.second", "id",
					),

					resource.TestCheckResourceAttr("data.scaleway_block_snapshots.by_tags_ordered_desc", "snapshots.#", "2"),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.second", "id", "data.scaleway_block_snapshots.by_tags_ordered_desc", "snapshots.0.id",
					),
					resource.TestCheckResourceAttrPair(
						"scaleway_block_snapshot.main", "id", "data.scaleway_block_snapshots.by_tags_ordered_desc", "snapshots.1.id",
					),

					resource.TestCheckResourceAttr("data.scaleway_block_snapshots.by_name_empty", "snapshots.#", "0"),
				),
			},
		},
	})
}
