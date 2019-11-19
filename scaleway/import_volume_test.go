package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewayVolume_importBasic(t *testing.T) {
	resourceName := "scaleway_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayVolumeConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
