package instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccImage_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isImageDestroyed(tt),
			instancechecks.IsSnapshotDestroyed(tt),
			instancechecks.IsVolumeDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_block_volume" "main" {
						size_in_gb = 20
						iops = 5000
					}

					resource "scaleway_block_snapshot" "main" {
						volume_id = scaleway_block_volume.main.id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 	= scaleway_block_snapshot.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volume_ids.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "state", "available"),
				),
			},
			{
				Config: `
					resource "scaleway_block_volume" "main" {
						size_in_gb = 20
						iops = 5000
					}

					resource "scaleway_block_snapshot" "main" {
						volume_id = scaleway_block_volume.main.id
					}

					resource "scaleway_instance_image" "main" {
						name 			= "tf-test-image-basic"
						root_volume_id 	= scaleway_block_snapshot.main.id
						tags 			= [ "add", "tags", "to-be", "removed", "later" ]
						architecture	= "arm64"
						public			= true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "name", "tf-test-image-basic"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "arm64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volume_ids.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.#", "5"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.0", "add"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.1", "tags"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.2", "to-be"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.3", "removed"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.4", "later"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "state", "available"),
				),
			},
			{
				Config: `
					resource "scaleway_block_volume" "main" {
						size_in_gb = 20
						iops = 5000
					}

					resource "scaleway_block_snapshot" "main" {
						volume_id = scaleway_block_volume.main.id
					}

					resource "scaleway_instance_image" "main" {
						name 			= "tf-test-image-basic-renamed"
						root_volume_id 	= scaleway_block_snapshot.main.id
						tags 			= [ "tags", "removed" ]
						architecture	= "x86_64"
						public			= false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "name", "tf-test-image-basic-renamed"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volume_ids.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.0", "tags"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "tags.1", "removed"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "state", "available"),
				),
			},
		},
	})
}

func TestAccImage_ExternalBlockVolume(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isImageDestroyed(tt),
			instancechecks.IsSnapshotDestroyed(tt),
			instancechecks.IsVolumeDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_block_volume" "main" {
						size_in_gb = 20
						iops = 5000
					}
					resource "scaleway_block_volume" "additional1" {
						size_in_gb = 20
						iops = 15000
					}
					resource "scaleway_block_volume" "additional2" {
						size_in_gb = 40
						iops = 5000
					}

					resource "scaleway_block_snapshot" "main" {
						volume_id = scaleway_block_volume.main.id
					}
					resource "scaleway_block_snapshot" "additional1" {
						volume_id = scaleway_block_volume.additional1.id
					}
					resource "scaleway_block_snapshot" "additional2" {
						volume_id = scaleway_block_volume.additional2.id
					}

					resource "scaleway_instance_image" "main" {
						name 			= "tf-test-image-external-block-volume"
						root_volume_id 	= scaleway_block_snapshot.main.id
						additional_volume_ids = [
							scaleway_block_snapshot.additional1.id,
							scaleway_block_snapshot.additional2.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volume_ids.#", "2"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volume_ids.0", "scaleway_block_snapshot.additional1", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volume_ids.1", "scaleway_block_snapshot.additional2", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_block_volume" "main" {
						size_in_gb = 20
						iops = 5000
					}
					resource "scaleway_block_volume" "additional1" {
						size_in_gb = 20
						iops = 15000
					}
					resource "scaleway_block_volume" "additional2" {
						size_in_gb = 40
						iops = 5000
					}

					resource "scaleway_block_snapshot" "main" {
						volume_id = scaleway_block_volume.main.id
					}
					resource "scaleway_block_snapshot" "additional1" {
						volume_id = scaleway_block_volume.additional1.id
					}
					resource "scaleway_block_snapshot" "additional2" {
						volume_id = scaleway_block_volume.additional2.id
					}

					resource "scaleway_instance_image" "main" {
						name 			= "tf-test-image-external-block-volume"
						root_volume_id 	= scaleway_block_snapshot.main.id
						additional_volume_ids = []
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_block_snapshot.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volume_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccImage_ServerWithLocalVolume(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isImageDestroyed(tt),
			instancechecks.IsSnapshotDestroyed(tt),
			instancechecks.IsVolumeDestroyed(tt),
			instancechecks.IsServerDestroyed(tt),
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
					resource "scaleway_instance_server" "server02" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							size_in_gb = 10
							volume_type = "l_ssd"
						}
					}
					resource "scaleway_instance_server" "server03" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							size_in_gb = 20
							volume_type = "l_ssd"
						}
					}

					resource "scaleway_instance_snapshot" "local01" {
						name = "snap01"
						volume_id = scaleway_instance_server.server01.root_volume.0.volume_id
					}
					resource "scaleway_instance_snapshot" "local02" {
						name = "snap02"
						volume_id = scaleway_instance_server.server02.root_volume.0.volume_id
					}
					resource "scaleway_instance_snapshot" "local03" {
						name = "snap03"
						volume_id = scaleway_instance_server.server03.root_volume.0.volume_id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.server01"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.server02"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.server03"),
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.local01"),
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.local02"),
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.local03"),
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
					resource "scaleway_instance_server" "server03" {
						image	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							size_in_gb = 20
							volume_type = "l_ssd"
						}
					}

					resource "scaleway_instance_snapshot" "local01" {
						name = "snap01"
						volume_id = scaleway_instance_server.server01.root_volume.0.volume_id
					}
					resource "scaleway_instance_snapshot" "local02" {
						name = "snap02"
						volume_id = scaleway_instance_server.server02.root_volume.0.volume_id
					}
					resource "scaleway_instance_snapshot" "local03" {
						name = "snap03"
						volume_id = scaleway_instance_server.server03.root_volume.0.volume_id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 	= scaleway_instance_snapshot.local01.id
						additional_volume_ids = [
							scaleway_instance_snapshot.local02.id,
							scaleway_instance_snapshot.local03.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.local01"),
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.local02"),
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.local03"),
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.local01", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume.0.id", "scaleway_instance_snapshot.local01", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "root_volume.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "root_volume.0.size", "15000000000"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volume_ids.0", "scaleway_instance_snapshot.local02", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_instance_snapshot.local02", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.size", "10000000000"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volume_ids.1", "scaleway_instance_snapshot.local03", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.1.id", "scaleway_instance_snapshot.local03", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.1.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.1.size", "20000000000"),
				),
			},
		},
	})
}

func TestAccImage_ServerWithSBSVolume(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isImageDestroyed(tt),
			blocktestfuncs.IsSnapshotDestroyed(tt),
			blocktestfuncs.IsVolumeDestroyed(tt),
			instancechecks.IsServerDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_block_volume" "block01" {
						iops       = 5000
						size_in_gb = 21
					}
					resource "scaleway_block_snapshot" "block01" {
						volume_id	= scaleway_block_volume.block01.id
					}

					resource "scaleway_instance_server" "server" {
						image	= "ubuntu_focal"
						type 	= "PLAY2-PICO"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.block01"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.server"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.block01"),
				),
			},
			{
				Config: `
					resource "scaleway_block_volume" "block01" {
						iops       = 5000
						size_in_gb = 21
					}
					resource "scaleway_block_snapshot" "block01" {
						volume_id	= scaleway_block_volume.block01.id
					}

					resource "scaleway_instance_server" "server" {
						image	= "ubuntu_focal"
						type 	= "PLAY2-PICO"
					}
					resource "scaleway_block_snapshot" "server" {
						volume_id 	= scaleway_instance_server.server.root_volume.0.volume_id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 	= scaleway_block_snapshot.server.id
						additional_volume_ids = [
							scaleway_block_snapshot.block01.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.block01"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.server"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.block01"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.server"),
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_block_snapshot.server", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_block_snapshot.block01", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.volume_type", "sbs_snapshot"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
				),
			},
			{
				Config: `
					resource "scaleway_block_volume" "block01" {
						iops       = 5000
						size_in_gb = 21
					}
					resource "scaleway_block_snapshot" "block01" {
						volume_id	= scaleway_block_volume.block01.id
					}

					resource "scaleway_block_volume" "block02" {
						iops       = 15000
						size_in_gb = 22
					}
					resource "scaleway_block_snapshot" "block02" {
						volume_id	= scaleway_block_volume.block02.id
					}

					resource "scaleway_instance_server" "server" {
						image	= "ubuntu_focal"
						type 	= "PLAY2-PICO"
					}
					resource "scaleway_block_snapshot" "server" {
						volume_id 	= scaleway_instance_server.server.root_volume.0.volume_id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 	= scaleway_block_snapshot.server.id
						additional_volume_ids = [
							scaleway_block_snapshot.block02.id,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.block01"),
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.block02"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.server"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.block01"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.block02"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.server"),
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_block_snapshot.server", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_block_snapshot.block02", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "additional_volumes.0.volume_type", "sbs_snapshot"),
					resource.TestCheckResourceAttrSet("scaleway_instance_image.main", "name"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("scaleway_instance_image.main", "public", "false"),
				),
			},
		},
	})
}

func TestAccImage_MixedVolumes(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	imageID := ""

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isImageDestroyed(tt),
			blocktestfuncs.IsSnapshotDestroyed(tt),
			instancechecks.IsSnapshotDestroyed(tt),
			instancechecks.IsServerDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "local" {
						image 	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							volume_type = "l_ssd"
						}
					}
					resource "scaleway_instance_server" "block_from_server" {
						image 	= "ubuntu_focal"
						type 	= "DEV1-S"
					}
					resource "scaleway_block_volume" "block_detached" {
						size_in_gb = 20
						iops = 5000
					}

					resource "scaleway_instance_snapshot" "local" {
						volume_id	= scaleway_instance_server.local.root_volume.0.volume_id
					}
					resource "scaleway_block_snapshot" "block_from_server" {
						volume_id	= scaleway_instance_server.block_from_server.root_volume.0.volume_id
					}
					resource "scaleway_block_snapshot" "block_detached" {
						volume_id	= scaleway_block_volume.block_detached.id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 			= scaleway_block_snapshot.block_detached.id
						additional_volume_ids	= [
							scaleway_block_snapshot.block_from_server.id,
							scaleway_instance_snapshot.local.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.block_from_server"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.block_detached"),
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.local"),
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_block_snapshot.block_detached", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_block_snapshot.block_from_server", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.1.id", "scaleway_instance_snapshot.local", "id"),
					acctest.CheckResourceIDPersisted("scaleway_instance_image.main", &imageID),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "local" {
						image 	= "ubuntu_focal"
						type 	= "DEV1-S"
						root_volume {
							volume_type = "l_ssd"
						}
					}
					resource "scaleway_instance_server" "block_from_server" {
						image 	= "ubuntu_focal"
						type 	= "DEV1-S"
					}
					resource "scaleway_block_volume" "block_detached" {
						size_in_gb = 20
						iops = 5000
					}

					resource "scaleway_instance_snapshot" "local" {
						volume_id	= scaleway_instance_server.local.root_volume.0.volume_id
					}
					resource "scaleway_block_snapshot" "block_from_server" {
						volume_id	= scaleway_instance_server.block_from_server.root_volume.0.volume_id
					}
					resource "scaleway_block_snapshot" "block_detached" {
						volume_id	= scaleway_block_volume.block_detached.id
					}

					resource "scaleway_instance_image" "main" {
						root_volume_id 			= scaleway_instance_snapshot.local.id
						additional_volume_ids	= [
							scaleway_block_snapshot.block_from_server.id,
							scaleway_block_snapshot.block_detached.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.block_from_server"),
					blocktestfuncs.IsSnapshotPresent(tt, "scaleway_block_snapshot.block_detached"),
					instancechecks.IsSnapshotPresent(tt, "scaleway_instance_snapshot.local"),
					instancechecks.DoesImageExists(tt, "scaleway_instance_image.main"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "root_volume_id", "scaleway_instance_snapshot.local", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.0.id", "scaleway_block_snapshot.block_from_server", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_image.main", "additional_volumes.1.id", "scaleway_block_snapshot.block_detached", "id"),
					acctest.CheckResourceIDChanged("scaleway_instance_image.main", &imageID),
				),
			},
		},
	})
}

func isImageDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_instance_image" {
				continue
			}

			instanceAPI, zone, ID, err := instance.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = instanceAPI.GetImage(&instanceSDK.GetImageRequest{
				ImageID: ID,
				Zone:    zone,
			})
			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("image (%s) still exists", rs.Primary.ID)
			}
			// Unexpected api error we return it
			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
