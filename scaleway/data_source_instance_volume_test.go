package scaleway

import (
	"fmt"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewayDataSourceInstanceVolume_Basic(t *testing.T) {
	dataSourceName := "data.scaleway_instance_volume.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayDataSourceVolumeConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayVolumeExists(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "size_in_gb", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "type", "l_ssd"),
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
`, rInt)
}
