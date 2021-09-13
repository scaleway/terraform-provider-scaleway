package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		CheckDestroy:      testAccCheckScalewayInstanceVolumeDestroy(tt),
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
					}`,
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
			},
		},
	})
}
