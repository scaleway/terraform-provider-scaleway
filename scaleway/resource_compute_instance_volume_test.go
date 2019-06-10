package scaleway

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("scaleway_compute_instance_volume", &resource.Sweeper{
		Name: "scaleway_compute_instance_volume",
		F:    testSweepComputeInstanceVolume,
	})
}

func testSweepComputeInstanceVolume(region string) error {

	// TODO: use new SDK

	scaleway, err := sharedDeprecatedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	log.Printf("[DEBUG] Destroying the volumes in (%s)", region)

	volumes, err := scaleway.GetVolumes()
	if err != nil {
		return fmt.Errorf("Error describing volumes in Sweeper: %s", err)
	}

	for _, volume := range *volumes {
		if err := scaleway.DeleteVolume(volume.Identifier); err != nil {
			return fmt.Errorf("Error deleting volume in Sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayComputeInstanceVolume_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceVolumeConfig,
				Check:  resource.ComposeTestCheckFunc(
				//testAccCheckScalewayComputeInstanceVolumeExists("scaleway_volume.test"),
				//testAccCheckScalewayComputeInstanceVolumeAttributes("scaleway_volume.test"),
				),
			},
		},
	})
}

func testAccCheckScalewayComputeInstanceVolumeDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Meta).deprecatedClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		_, err := client.GetVolume(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Volume still exists")
		}
	}

	return nil
}

var testAccCheckScalewayComputeInstanceVolumeConfig = `
resource "scaleway_compute_instance_volume" "test" {
  name = "terraform-test"
  size_in_gb = 2
  type = "l_ssd"
}
`
