package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceInstanceVolume_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	volumeName := "tf-volume"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceVolumeDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_volume" "test" {
						name = "%s"
						size_in_gb = 2
						type = "l_ssd"
					}`, volumeName),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_volume" "test" {
						name = "%s"
						size_in_gb = 2
						type = "l_ssd"
					}
					
					data "scaleway_instance_volume" "test" {
						name = "${scaleway_instance_volume.test.name}"
					}
					
					data "scaleway_instance_volume" "test2" {
						volume_id = "${scaleway_instance_volume.test.id}"
					}
				`, volumeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists(tt, "data.scaleway_instance_volume.test"),
					resource.TestCheckResourceAttr("data.scaleway_instance_volume.test", "size_in_gb", "2"),
					resource.TestCheckResourceAttr("data.scaleway_instance_volume.test", "type", "l_ssd"),
				),
			},
		},
	})
}
