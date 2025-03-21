package block_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestAccSnapshot_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      blocktestfuncs.IsSnapshotDestroyed(tt),
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
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.main"),
					acctest.CheckResourceAttrUUID("scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_block_snapshot.main", "name", "test-block-snapshot-basic"),
				),
			},
		},
	})
}
