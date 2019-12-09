package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_volume", &resource.Sweeper{
		Name: "scaleway_instance_volume",
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

func TestAccSalewayInstanceVolume_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSalewayInstanceVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSalewayInstanceVolumeConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.test"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.test", "size_in_gb", "20"),
				),
			},
			{
				Config: testAccCheckSalewayInstanceVolumeConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.test"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.test", "name", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.test", "size_in_gb", "20"),
				),
			},
		},
	})
}

func TestAccSalewayInstanceVolume_FromVolume(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSalewayInstanceVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSalewayInstanceVolumeConfigFromVolume,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.test1"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.test2"),
				),
			},
		},
	})
}

func TestAccSalewayInstanceVolume_RandomName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSalewayInstanceVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSalewayInstanceVolumeConfigWithRandomName[0],
			},
			{
				Config: testAccCheckSalewayInstanceVolumeConfigWithRandomName[1],
			},
		},
	})
}

func testAccCheckScalewayInstanceVolumeExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
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

func testAccCheckSalewayInstanceVolumeDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta().(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_instance_volume" {
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

var testAccCheckSalewayInstanceVolumeConfig = []string{
	`
		resource "scaleway_instance_volume" "test" {
			type       = "l_ssd"
			size_in_gb = 20
		}
	`,
	`
		resource "scaleway_instance_volume" "test" {
			type       = "l_ssd"
			name       = "terraform-test"
			size_in_gb = 20
		}
	`,
}

var testAccCheckSalewayInstanceVolumeConfigWithRandomName = []string{
	`
		resource "scaleway_instance_volume" "test" {
			type       = "l_ssd"
			size_in_gb = 20
		}
	`,
	`
		resource "scaleway_instance_volume" "test" {
			type       = "l_ssd"
			size_in_gb = 20
		}
	`,
}

var testAccCheckSalewayInstanceVolumeConfigFromVolume = `
		resource "scaleway_instance_volume" "test1" {
			type       = "l_ssd"
			size_in_gb = 20
		}

		resource "scaleway_instance_volume" "test2" {
			type           = "l_ssd"
			from_volume_id = "${scaleway_instance_volume.test1.id}"
		}
	`
