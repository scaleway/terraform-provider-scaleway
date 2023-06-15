package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceBlockVolume_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayBlockVolumeDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_block_volume main {
						type = "b_ssd"
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
					testAccCheckScalewayBlockVolumeExists(tt, "scaleway_block_volume.main"),

					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "name", "data.scaleway_block_volume.find_by_name", "name"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "name", "data.scaleway_block_volume.find_by_id", "name"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "id", "data.scaleway_block_volume.find_by_name", "id"),
					resource.TestCheckResourceAttrPair("scaleway_block_volume.main", "id", "data.scaleway_block_volume.find_by_id", "id"),
				),
			},
		},
	})
}
