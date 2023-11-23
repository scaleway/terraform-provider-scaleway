package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceBlockSnapshot_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayBlockSnapshotDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						iops = 5000
						size_in_gb = 10
					}

					resource scaleway_block_snapshot main {
						name = "test-ds-block-snapshot-basic"
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
					testAccCheckScalewayBlockSnapshotExists(tt, "scaleway_block_snapshot.main"),

					resource.TestCheckResourceAttrPair("scaleway_block_snapshot.main", "name", "data.scaleway_block_snapshot.find_by_name", "name"),
					resource.TestCheckResourceAttrPair("scaleway_block_snapshot.main", "name", "data.scaleway_block_snapshot.find_by_id", "name"),
					resource.TestCheckResourceAttrPair("scaleway_block_snapshot.main", "id", "data.scaleway_block_snapshot.find_by_name", "id"),
					resource.TestCheckResourceAttrPair("scaleway_block_snapshot.main", "id", "data.scaleway_block_snapshot.find_by_id", "id"),
				),
			},
		},
	})
}
