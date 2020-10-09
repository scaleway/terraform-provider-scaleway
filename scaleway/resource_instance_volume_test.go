package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_volume", &resource.Sweeper{
		Name: "scaleway_instance_volume",
		F:    testSweepComputeInstanceVolume,
	})
}

func testSweepComputeInstanceVolume(region string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client) error {
		instanceAPI := instance.NewAPI(scwClient)
		zone, _ := scwClient.GetDefaultZone()
		l.Debugf("sweeper: destroying the volumes in (%s)", zone)

		listVolumesResponse, err := instanceAPI.ListVolumes(&instance.ListVolumesRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing volumes in sweeper: %s", err)
		}

		for _, volume := range listVolumesResponse.Volumes {
			if volume.Server == nil {
				err := instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
					Zone:     zone,
					VolumeID: volume.ID,
				})
				if err != nil {
					return fmt.Errorf("error deleting volume in sweeper: %s", err)
				}
			}
		}

		return nil
	})
}

func TestAccScalewayInstanceVolume_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceVolumeConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.test"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.test", "size_in_gb", "20"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceVolumeConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.test"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.test", "name", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.test", "size_in_gb", "20"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceVolume_FromVolume(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceVolumeConfigFromVolume,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.test1"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.test2"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceVolume_RandomName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceVolumeConfigWithRandomName[0],
			},
			{
				Config: testAccCheckScalewayInstanceVolumeConfigWithRandomName[1],
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

func testAccCheckScalewayInstanceVolumeDestroy(s *terraform.State) error {
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

var testAccCheckScalewayInstanceVolumeConfig = []string{
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

var testAccCheckScalewayInstanceVolumeConfigWithRandomName = []string{
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

var testAccCheckScalewayInstanceVolumeConfigFromVolume = `
		resource "scaleway_instance_volume" "test1" {
			type       = "l_ssd"
			size_in_gb = 20
		}

		resource "scaleway_instance_volume" "test2" {
			type           = "l_ssd"
			from_volume_id = "${scaleway_instance_volume.test1.id}"
		}
	`
