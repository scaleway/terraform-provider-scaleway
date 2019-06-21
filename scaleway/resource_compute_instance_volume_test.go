package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

var testAccCheckScalewayComputeInstanceVolumeConfig = []string{
	`
		resource "scaleway_compute_instance_volume" "test" {
			size = "20GB"
		}
	`,
	`
		resource "scaleway_compute_instance_volume" "test" {
			name = "terraform-test"
			size = "20000000000"
		}
	`,
}

var testAccCheckScalewayComputeInstanceVolumeConfigWithRandomName = []string{
	`
		resource "scaleway_compute_instance_volume" "test" {
			size = "20GB"
		}
	`,
	`
		resource "scaleway_compute_instance_volume" "test" {
			size = "20000000000"
		}
	`,
}

func init() {
	resource.AddTestSweepers("scaleway_compute_instance_volume", &resource.Sweeper{
		Name: "scaleway_compute_instance_volume",
		F:    testSweepComputeInstanceVolume,
	})
}

func testSweepComputeInstanceVolume(region string) error {
	scwClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	instanceAPI := instance.NewAPI(scwClient)

	l.Debugf("sweeper: destroying the volumes in (%s)", region)

	listVolumesResponse, err := instanceAPI.ListVolumes(&instance.ListVolumesRequest{})
	if err != nil {
		return fmt.Errorf("error listing volumes in sweeper: %s", err)
	}

	for _, volume := range listVolumesResponse.Volumes {
		err := instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
			VolumeID: volume.ID,
		})
		if err != nil {
			return fmt.Errorf("error deleting volume in sweeper: %s", err)
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
				Config: testAccCheckScalewayComputeInstanceVolumeConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.test", "size", "20GB"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceVolumeConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.test", "name", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.test", "size", "20GB"),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceVolume_RandomName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceVolumeConfigWithRandomName[0],
			},
			{
				Config: testAccCheckScalewayComputeInstanceVolumeConfigWithRandomName[1],
			},
		},
	})
}

func testAccCheckScalewayComputeInstanceVolumeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		zone, id, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}

		meta := testAccProvider.Meta().(*Meta)
		instanceAPI := instance.NewAPI(meta.scwClient)
		_, err = instanceAPI.GetVolume(&instance.GetVolumeRequest{
			VolumeID: id,
			Zone:     zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayComputeInstanceVolumeDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta().(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_compute_instance_volume" {
			continue
		}

		zone, id, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetVolume(&instance.GetVolumeRequest{
			Zone:     zone,
			VolumeID: id,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("volume (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}
