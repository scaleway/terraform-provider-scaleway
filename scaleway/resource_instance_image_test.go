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
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_volume.main.id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id = scaleway_instance_snapshot.main.id
						architecture = "x86_64"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
				),
			},
		},
	})
}

//func TestAccScalewayInstanceImage_Server(t *testing.T) {
//	tt := NewTestTools(t)
//	defer tt.Cleanup()
//	resource.Test(t, resource.TestCase{
//		PreCheck:          func() { testAccPreCheck(t) },
//		ProviderFactories: tt.ProviderFactories,
//		CheckDestroy: resource.ComposeTestCheckFunc(
//			testAccCheckScalewayInstanceImageDestroy(tt),
//			testAccCheckScalewayInstanceServerDestroy(tt),
//		),
//		Steps: []resource.TestStep{
//			{
//				Config: `
//					resource "scaleway_instance_server" "main" {
//						image = "ubuntu_focal"
//						type = "DEV1-S"
//						state = "stopped"
//					}
//
//					resource "scaleway_instance_image" "main" {
//						name = "test_image_basic"
//						root_volume_id = scaleway_instance_server.main.root_volume.0.volume_id
//						architecture = "arm"
//					}
//				`,
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckScalewayInstanceImageExists(tt, "scaleway_instance_image.main"),
//					resource.TestCheckResourceAttr("scaleway_instance_image.main", "name", "test_image_basic"),
//					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "arm"),
//					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_server.main", "root_volume.0.volume_id"),
//				),
//			},
//		},
//	})
//}

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
