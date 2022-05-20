package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func TestAccScalewayInstanceSnapshot_BlockVolume(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceVolumeDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "main" {
						type       = "b_ssd"
						size_in_gb = 20
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_volume.main.id
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.main"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceSnapshot_Server(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceVolumeDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						image = "ubuntu_focal"
						type = "DEV1-S"
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.main"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceSnapshot_ServerWithBlockVolume(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayInstanceVolumeDestroy(tt),
			testAccCheckScalewayInstanceServerDestroy(tt),
			testAccCheckScalewayInstanceSnapshotDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "block" {
						type       = "b_ssd"
						size_in_gb = 10
					}

					resource "scaleway_instance_server" "main" {
						image = "ubuntu_focal"
						type = "DEV1-S"

						additional_volume_ids = [
							scaleway_instance_volume.block.id
						]
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_volume.block.id
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.main"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceSnapshot_RenameSnapshot(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceVolumeDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "main" {
						type       = "b_ssd"
						size_in_gb = 20
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_volume.main.id
						name = "first_name"
						tags = ["test-terraform"]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.main"),
					resource.TestCheckResourceAttr("scaleway_instance_snapshot.main", "tags.0", "test-terraform"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_volume" "main" {
						type       = "b_ssd"
						size_in_gb = 20
					}

					resource "scaleway_instance_snapshot" "main" {
						volume_id = scaleway_instance_volume.main.id
						name = "second_name"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceSnapShotExists(tt, "scaleway_instance_snapshot.main"),
					resource.TestCheckResourceAttr("scaleway_instance_snapshot.main", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceSnapShotExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetSnapshot(&instance.GetSnapshotRequest{
			Zone:       zone,
			SnapshotID: ID,
		})

		if err != nil {
			return fmt.Errorf("failed to find snapshot %s: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckScalewayInstanceSnapshotDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_instance_snapshot" {
				continue
			}

			instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = instanceAPI.GetSnapshot(&instance.GetSnapshotRequest{
				SnapshotID: ID,
				Zone:       zone,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("snapshot (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return fmt.Errorf("unexpected error when testing snapshot (%s) : %s", rs.Primary.ID, err)
			}
		}

		return nil
	}
}
