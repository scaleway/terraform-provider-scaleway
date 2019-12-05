package scaleway

import (
	"fmt"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewayDataSourceInstanceVolume_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSalewayInstanceVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayDataSourceInstanceVolumeConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("data.scaleway_instance_volume.test"),
					resource.TestCheckResourceAttr("data.scaleway_instance_volume.test", "size_in_gb", "2"),
					resource.TestCheckResourceAttr("data.scaleway_instance_volume.test", "type", "l_ssd"),
				),
			},
		},
	})
}

func testAccCheckScalewayDataSourceInstanceVolumeConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scaleway_instance_volume" "test" {
  name = "test%d"
  size_in_gb = 2
  type = "l_ssd"
}

data "scaleway_instance_volume" "test" {
  name = "${scaleway_instance_volume.test.name}"
}

data "scaleway_instance_volume" "test2" {
  volume_id = "${scaleway_instance_volume.test.id}"
}
`, rInt)
}
