package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewayDataSourceInstanceVolume_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "test" {
						name = "` + newRandomName("volume") + `"
						size_in_gb = 2
						type = "l_ssd"
					}
					
					data "scaleway_instance_volume" "test" {
						name = "${scaleway_instance_volume.test.name}"
					}
					
					data "scaleway_instance_volume" "test2" {
						volume_id = "${scaleway_instance_volume.test.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("data.scaleway_instance_volume.test"),
					resource.TestCheckResourceAttr("data.scaleway_instance_volume.test", "size_in_gb", "2"),
					resource.TestCheckResourceAttr("data.scaleway_instance_volume.test", "type", "l_ssd"),
				),
			},
		},
	})
}
