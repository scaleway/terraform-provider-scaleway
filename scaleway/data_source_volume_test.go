package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalewayDataSourceVolume_Basic(t *testing.T) {
	t.Parallel()

	dataSourceName := "data.scaleway_volume.test"
	resource.Test(t, resource.TestCase{
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

func testAccCheckScalewayDataSourceVolumeConfig(rInt int) string {
	return fmt.Sprintf(`
resource "scaleway_volume" "test" {
  name = "test%d"
  size_in_gb = 2
  type = "l_ssd"
}

data "scaleway_volume" "test" {
  name = "${scaleway_volume.test.name}"
}
`, rInt)
}
