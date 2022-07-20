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
	resource.ParallelTest(t, resource.TestCase{
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
						type       	= "b_ssd"
						size_in_gb 	= 20
						zone 		= "nl-ams-1"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id 	= scaleway_instance_volume.main.id
						zone 		= "nl-ams-1"
					}

					resource "scaleway_instance_image" "main" {
						name 			= "test_image_basic"
						root_volume_id 	= scaleway_instance_snapshot.main.id
						tags 			= ["tag1", "tag2", "tag3"]
						zone 			= "nl-ams-1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists(tt, "scaleway_instance_volume.main"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.main"),
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "name", "test_image_basic"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.1", "tag2"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.2", "tag3"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "zone", "nl-ams-1"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "creation_date"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "modification_date"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "modification_date", "scaleway_instance_image.main", "creation_date"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "state"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "organization_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "project_id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_volume" "main" {
						type       	= "b_ssd"
						size_in_gb 	= 20
						zone 		= "nl-ams-1"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id 	= scaleway_instance_volume.main.id
						zone 		= "nl-ams-1"
					}

					resource "scaleway_instance_image" "main" {
						name 			= "test_image_renamed"
						root_volume_id 	= scaleway_instance_snapshot.main.id
						tags 			= ["new tag"]
						public 			= true
						architecture	= "arm"
						zone 			= "nl-ams-1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists(tt, "scaleway_instance_volume.main"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.main"),
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "name", "test_image_renamed"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.0", "new tag"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "arm"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "zone", "nl-ams-1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "creation_date"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "modification_date"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "state"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "organization_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "project_id"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceImage_Server(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
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
						image 	= "ubuntu_focal"
						type 	= "DEV1-S"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id	= scaleway_instance_server.main.root_volume.0.volume_id
						depends_on 	= [ scaleway_instance_server.main ]
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 			= scaleway_instance_snapshot.main.id
						tags 					= ["test_remove_tags"]
						depends_on 				= [ scaleway_instance_snapshot.main ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.main"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.main"),
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.0", "test_remove_tags"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						image 	= "ubuntu_focal"
						type 	= "DEV1-S"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id 	= scaleway_instance_server.main.root_volume.0.volume_id
						depends_on	= [ scaleway_instance_server.main ]
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 			= scaleway_instance_snapshot.main.id
						depends_on 				= [ scaleway_instance_snapshot.main ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.main"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.main"),
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_instance_image.main", "tags.0"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceImage_ServerWithBlockVolume(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
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
					resource "scaleway_instance_volume" "block01" {
						type       = "b_ssd"
						size_in_gb = 21
					}
					resource "scaleway_instance_snapshot" "block01" {
						volume_id	= scaleway_instance_volume.block01.id
						depends_on 	= [ scaleway_instance_volume.block01 ]
					}

					resource "scaleway_instance_server" "server" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists(tt, "scaleway_instance_volume.block01"),
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.block01"),
				),
			},

			{
				Config: `
					resource "scaleway_instance_volume" "block01" {
						type       = "b_ssd"
						size_in_gb = 21
					}
					resource "scaleway_instance_snapshot" "block01" {
						volume_id	= scaleway_instance_volume.block01.id
						depends_on 	= [ scaleway_instance_volume.block01 ]
					}

					resource "scaleway_instance_server" "server" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
					}
					resource "scaleway_instance_snapshot" "server" {
						volume_id 	= scaleway_instance_server.server.root_volume.0.volume_id
						depends_on 	= [ scaleway_instance_server.server ]
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 	= scaleway_instance_snapshot.server.id
						additional_volume_ids = [
							scaleway_instance_snapshot.block01.id
						]
						depends_on = [
							scaleway_instance_snapshot.block01,
							scaleway_instance_snapshot.server,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists(tt, "scaleway_instance_volume.block01"),
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.block01"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.server"),
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.server", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_instance_snapshot.block01", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.volume_type", "b_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.size", "21000000000"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_volume" "block01" {
						type       = "b_ssd"
						size_in_gb = 21
					}
					resource "scaleway_instance_snapshot" "block01" {
						volume_id	= scaleway_instance_volume.block01.id
						depends_on 	= [ scaleway_instance_volume.block01 ]
					}

					resource "scaleway_instance_volume" "block02" {
						type       = "b_ssd"
						size_in_gb = 22
					}
					resource "scaleway_instance_snapshot" "block02" {
						volume_id	= scaleway_instance_volume.block02.id
						depends_on 	= [ scaleway_instance_volume.block02 ]
					}

					resource "scaleway_instance_server" "server" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
					}
					resource "scaleway_instance_snapshot" "server" {
						volume_id 	= scaleway_instance_server.server.root_volume.0.volume_id
						depends_on 	= [ scaleway_instance_server.server ]
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 	= scaleway_instance_snapshot.server.id
						additional_volume_ids = [
							scaleway_instance_snapshot.block02.id,
						]
						depends_on = [
							scaleway_instance_snapshot.block02,
							scaleway_instance_snapshot.server,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists(tt, "scaleway_instance_volume.block01"),
					testAccCheckScalewayInstanceVolumeExists(tt, "scaleway_instance_volume.block02"),
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.block01"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.block02"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.server"),
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.server", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_instance_snapshot.block02", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.volume_type", "b_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.size", "22000000000"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceImage_ServerWithLocalVolume(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
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
					resource "scaleway_instance_server" "server01" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							size_in_gb = 15
							volume_type = "l_ssd"
						}
					}
					resource "scaleway_instance_snapshot" "local01" {
						volume_id = scaleway_instance_server.server01.root_volume.0.volume_id
						depends_on = [ scaleway_instance_server.server01 ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server01"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.local01"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "server01" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							size_in_gb = 15
							volume_type = "l_ssd"
						}
					}
					resource "scaleway_instance_server" "server02" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							size_in_gb = 10
							volume_type = "l_ssd"
						}
					}

					resource "scaleway_instance_snapshot" "local01" {
						volume_id = scaleway_instance_server.server01.root_volume.0.volume_id
						depends_on = [ scaleway_instance_server.server01 ]
					}
					resource "scaleway_instance_snapshot" "local02" {
						volume_id = scaleway_instance_server.server02.root_volume.0.volume_id
						depends_on = [ scaleway_instance_server.server02 ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server01"),
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server02"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.local01"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.local02"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "server01" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							size_in_gb = 15
							volume_type = "l_ssd"
						}
					}
					resource "scaleway_instance_server" "server02" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							size_in_gb = 10
							volume_type = "l_ssd"
						}
					}

					resource "scaleway_instance_snapshot" "local01" {
						volume_id = scaleway_instance_server.server01.root_volume.0.volume_id
						depends_on = [ scaleway_instance_server.server01 ]
					}
					resource "scaleway_instance_snapshot" "local02" {
						volume_id = scaleway_instance_server.server02.root_volume.0.volume_id
						depends_on = [ scaleway_instance_server.server02 ]
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 	= scaleway_instance_snapshot.local01.id
						additional_volume_ids = [ scaleway_instance_snapshot.local02.id ]
						depends_on = [
							scaleway_instance_snapshot.local01,
							scaleway_instance_snapshot.local02,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server01"),
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server02"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.local01"),
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.local02"),
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.local01", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_instance_snapshot.local02", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.size", "10000000000"),
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
