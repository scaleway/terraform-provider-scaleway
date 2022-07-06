package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func TestAccScalewayInstanceImage_BlockVolume(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayInstanceImageDestroy(tt),
			testAccCheckScalewayInstanceSnapshotDestroy(tt),
			testAccCheckScalewayInstanceVolumeDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "main" {
						type       = "b_ssd"
						size_in_gb = 20
						zone = "nl-ams-1"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_volume.main.id
						zone = "nl-ams-1"
					}

					resource "scaleway_instance_image" "main" {
						name = "test_image_basic"
						root_volume_id = scaleway_instance_snapshot.main.id
						architecture = "arm"
						tags = ["tag1", "tag2", "tag3"]
						public = true
						zone = "nl-ams-1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "name", "test_image_basic"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "arm"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "zone", "nl-ams-1"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.1", "tag2"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.2", "tag3"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "true"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "creation_date"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "modification_date"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "modification_date", "scaleway_instance_image.main", "creation_date"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "state"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "organization_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "project_id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_volume" "main" {
						type       = "b_ssd"
						size_in_gb = 20
						zone = "nl-ams-1"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_volume.main.id
						zone = "nl-ams-1"
					}

					resource "scaleway_instance_image" "main" {
						name = "test_image_renamed"
						root_volume_id = scaleway_instance_snapshot.main.id
						tags = ["tag3"]
						public = false
						zone = "nl-ams-1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "name", "test_image_renamed"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "zone", "nl-ams-1"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.0", "tag3"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "creation_date"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "modification_date"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "state"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "organization_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "project_id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceImage_Server(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayInstanceImageDestroy(tt),
			testAccCheckScalewayInstanceSnapshotDestroy(tt),
			testAccCheckScalewayInstanceServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						image = "ubuntu_focal"
						type = "DEV1-S"
						state = "stopped"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id = scaleway_instance_snapshot.main.id
						architecture = "x86_64"
						tags = ["tag"]
						default_bootscript_id = "eb760e3c-30d8-49a3-b3ad-ad10c3aa440b"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.0", "tag"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "default_bootscript.id", "eb760e3c-30d8-49a3-b3ad-ad10c3aa440b"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "default_bootscript.kernel", "http://10.194.3.9/kernel/x86_64-mainline-lts-4.14-latest/vmlinuz"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "default_bootscript.initrd", "http://10.194.3.9/initrd/initrd-Linux-x86_64-v3.14.6.gz"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "default_bootscript.bootcmdargs", "LINUX_COMMON scaleway boot=local nbd.max_part=16"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						image = "ubuntu_focal"
						type = "DEV1-S"
						state = "stopped"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id = scaleway_instance_snapshot.main.id
						architecture = "x86_64"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_instance_image.main", "tags.0"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceImage_ServerWithBlockVolume(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayInstanceImageDestroy(tt),
			testAccCheckScalewayInstanceSnapshotDestroy(tt),
			testAccCheckScalewayInstanceVolumeDestroy(tt),
			testAccCheckScalewayInstanceServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "block" {
						type       = "b_ssd"
						size_in_gb = 20
					}

					resource "scaleway_instance_server" "main" {
						image = "ubuntu_focal"
						type = "DEV1-S"
						state = "stopped"
						additional_volume_ids = [
							scaleway_instance_volume.block.id
						]
					}

					resource "scaleway_instance_snapshot" "block" {
						volume_id = scaleway_instance_volume.block.id
					}
					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}

					resource "scaleway_instance_image" "main" {
						architecture = "x86_64"
						root_volume_id = scaleway_instance_snapshot.main.id
						additional_volume_ids = [
							scaleway_instance_snapshot.block.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_instance_snapshot.block", "id"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceImage_ServerWithLocalVolume(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayInstanceImageDestroy(tt),
			testAccCheckScalewayInstanceSnapshotDestroy(tt),
			testAccCheckScalewayInstanceVolumeDestroy(tt),
			testAccCheckScalewayInstanceServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						image = "ubuntu_focal"
						type = "DEV1-S"
						state = "stopped"
					}

					resource "scaleway_instance_volume" "main" {
						type       = "l_ssd"
						size_in_gb = 20
						zone = "nl-ams-1"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id = scaleway_instance_snapshot.main.id
						architecture = "x86_64"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceImage_SeveralVolumes(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayInstanceImageDestroy(tt),
			testAccCheckScalewayInstanceSnapshotDestroy(tt),
			testAccCheckScalewayInstanceVolumeDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "vol00" {
						type       = "b_ssd"
						size_in_gb = 20
					}
					resource "scaleway_instance_snapshot" "snap00" {
						volume_id = scaleway_instance_volume.vol00.id
					}

					resource "scaleway_instance_volume" "vol01" {
						type       = "b_ssd"
						size_in_gb = 20
					}
					resource "scaleway_instance_snapshot" "snap01" {
						volume_id = scaleway_instance_volume.vol01.id
					}

					resource "scaleway_instance_volume" "vol02" {
						type       = "b_ssd"
						size_in_gb = 20
					}
					resource "scaleway_instance_snapshot" "snap02" {
						volume_id = scaleway_instance_volume.vol02.id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id = 	scaleway_instance_snapshot.snap00.id
						architecture = 		"x86_64"
						additional_volume_ids = [
							scaleway_instance_snapshot.snap01.id,
							scaleway_instance_snapshot.snap02.id,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.snap00", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_instance_snapshot.snap01", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.1.id", "scaleway_instance_snapshot.snap02", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceImageExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		zone, ID, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}
		instanceAPI := instance.NewAPI(tt.Meta.scwClient)
		_, err = instanceAPI.GetImage(&instance.GetImageRequest{
			ImageID: ID,
			Zone:    zone,
		})
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckScalewayInstanceImageDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_instance_image" {
				continue
			}
			instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}
			_, err = instanceAPI.GetImage(&instance.GetImageRequest{
				ImageID: ID,
				Zone:    zone,
			})
			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("image (%s) still exists", rs.Primary.ID)
			}
			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
