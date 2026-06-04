package instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccServer_RootVolume_Size(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" main {
					  name = "tf-acc-server-root-volume-size"
					  image = "ubuntu_focal"
					  type  = "DEV1-S"
					  root_volume {
						volume_type = "l_ssd"
					  }
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "name", "tf-acc-server-root-volume-size"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "image", "ubuntu_focal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.size_in_gb", "20"), // It resizes to 20GB as terraform will take max space available for l_ssd.
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "enable_dynamic_ip", "false"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" main2 {
					  name = "tf-acc-server-root-volume-size"
					  image = "ubuntu_focal"
					  type  = "DEV1-S"
					  root_volume {
						volume_type = "l_ssd"
						size_in_gb  = 20
					  }
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.main2"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main2", "name", "tf-acc-server-root-volume-size"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main2", "image", "ubuntu_focal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main2", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main2", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main2", "root_volume.0.size_in_gb", "20"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main2", "root_volume.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main2", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main2", "enable_dynamic_ip", "false"),
				),
			},
		},
	})
}

func TestAccServer_RootVolumeFromImage_Block(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	serverID := ""

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-root-volume-from-image-block"
						image = "%s"
						type  = "DEV1-S"
						root_volume {
							volume_type = "sbs_volume"
							size_in_gb = 10
							sbs_iops = 5000
							name = "named-volume"
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_image_block" ]
					}`, ubuntuFocalImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-root-volume-from-image-block"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.volume_type", "sbs_volume"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.sbs_iops", "5000"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.name", "named-volume"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", &serverID),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-root-volume-from-image-block"
						image = "%s"
						type  = "DEV1-S"
						root_volume {
							volume_type = "sbs_volume"
							size_in_gb = 20
							sbs_iops = 15000
							name = "renamed"
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_image_block" ]
					}`, ubuntuFocalImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-root-volume-from-image-block"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "20"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.volume_type", "sbs_volume"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.sbs_iops", "15000"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.name", "renamed"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", &serverID),
				),
			},
		},
	})
}

func TestAccServer_RootVolumeFromImage_Local(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	serverID := ""

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-root-volume-from-image-local"
						image = "%s"
						type  = "DEV1-S"
						root_volume {
							volume_type = "l_ssd"
							size_in_gb = 10
							name = "named-volume"
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_image_local" ]
					}`, ubuntuFocalImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-root-volume-from-image-local"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.name", "named-volume"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", &serverID),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-root-volume-from-image-local"
						image = "%s"
						type  = "DEV1-S"
						root_volume {
							volume_type = "l_ssd"
							size_in_gb = 10
							name = "renamed"
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_image_local" ]
					}`, ubuntuFocalImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-root-volume-from-image-local"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.name", "renamed"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", &serverID),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-root-volume-from-image-local"
						image = "%s"
						type  = "DEV1-S"
						root_volume {
							volume_type = "l_ssd"
							size_in_gb = 20
							delete_on_termination = true
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_image_local" ]
					}`, ubuntuFocalImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					serverHasNewVolume(tt, "scaleway_instance_server.base", ubuntuFocalImageLabel),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-root-volume-from-image-local"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "20"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "root_volume.0.name"),
					acctest.CheckResourceIDChanged("scaleway_instance_server.base", &serverID), // Server should have been re-created as l_ssd cannot be resized.
				),
			},
		},
	})
}

func TestAccServer_RootVolumeFromID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	serverID := ""

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "local" {
						type = "l_ssd"
						size_in_gb = 10
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-root-volume-from-id"
						type  = "DEV1-S"
						root_volume {
							volume_id = scaleway_instance_volume.local.id
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_id" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-root-volume-from-id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "root_volume.0.name"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "root_volume.0.volume_id", "scaleway_instance_volume.local", "id"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", &serverID),
				),
			},
			{
				Config: `
					resource "scaleway_instance_volume" "local" {
						type = "l_ssd"
						size_in_gb = 10
						name = "named-volume"
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-root-volume-from-id"
						type  = "DEV1-S"
						root_volume {
							volume_id = scaleway_instance_volume.local.id
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_id" ]
					}`,
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-root-volume-from-id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.name", "named-volume"), // New volume name should be readable from root_volume attribute after refreshing the state
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "root_volume.0.name", "scaleway_instance_volume.local", "name"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "root_volume.0.volume_id", "scaleway_instance_volume.local", "id"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", &serverID),
				),
			},
		},
	})
}

func TestAccServer_RootVolumeFromExternalSnapshot(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-root-volume-from-external-snapshot"
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						root_volume {
							volume_type = "sbs_volume"
							size_in_gb = 50
							sbs_iops = 5000
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_external_snapshot" ]
					}

					resource "scaleway_block_snapshot" "snapshot" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "name", "tf-acc-server-root-volume-from-external-snapshot"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "PLAY2-PICO"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "additional_volume_ids.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.volume_type", string(instanceSDK.VolumeVolumeTypeSbsVolume)),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.sbs_iops", "5000"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.size_in_gb", "50"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-root-volume-from-external-snapshot"
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						root_volume {
							volume_type = "sbs_volume"
							size_in_gb = 50
							sbs_iops = 5000
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_external_snapshot" ]
					}

					resource "scaleway_block_snapshot" "snapshot" {
						volume_id = scaleway_instance_server.main.root_volume.0.volume_id
					}

					resource "scaleway_block_volume" "volume" {
						snapshot_id = scaleway_block_snapshot.snapshot.id
						iops = 5000
					}

					resource "scaleway_instance_server" "from_snapshot" {
						name = "tf-acc-server-root-volume-from-external-snapshot-2"
						type  = "PLAY2-PICO"
						root_volume {
							volume_type = "sbs_volume"
							volume_id = scaleway_block_volume.volume.id
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_from_external_snapshot" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "name", "tf-acc-server-root-volume-from-external-snapshot"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "PLAY2-PICO"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "additional_volume_ids.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.volume_type", string(instanceSDK.VolumeVolumeTypeSbsVolume)),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.sbs_iops", "5000"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.size_in_gb", "50"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.from_snapshot", "root_volume.0.volume_id", "scaleway_block_volume.volume", "id"),
				),
			},
		},
	})
}

func TestAccServer_RootVolume_Boot(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-root-volume-boot"
						image = "%s"
						type  = "DEV1-S"
						state = "stopped"
						root_volume {
							boot = true
							delete_on_termination = true
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_boot" ]
					}`, ubuntuFocalImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-root-volume-boot"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.boot", "true"),
					serverHasNewVolume(tt, "scaleway_instance_server.base", "ubuntu_focal"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-root-volume-boot"
						image = "%s"
						type  = "DEV1-S"
						state = "stopped"
						root_volume {
							boot = false
							delete_on_termination = true
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume_boot" ]
					}`, ubuntuFocalImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-root-volume-boot"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.boot", "false"),
					serverHasNewVolume(tt, "scaleway_instance_server.base", "ubuntu_focal"),
				),
			},
		},
	})
}

func TestAccServer_AdditionalVolumes(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// With additional local
				Config: `
					resource "scaleway_instance_volume" "local" {
						size_in_gb = 10
						type = "l_ssd"
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-additional-volumes"
						image = "ubuntu_focal"
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "additional_volumes" ]

						root_volume {
							size_in_gb = 15
						}
						additional_volume_ids = [
							scaleway_instance_volume.local.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					instancechecks.IsVolumePresent(tt, "scaleway_instance_volume.local"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-additional-volumes"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "15"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "additional_volume_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "additional_volume_ids.0", "scaleway_instance_volume.local", "id"),
				),
			},
			{
				// With additional local and block
				Config: `
					resource "scaleway_instance_volume" "local" {
						size_in_gb = 10
						type = "l_ssd"
					}

					resource "scaleway_block_volume" "block" {
						size_in_gb = 12
						iops = 5000
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-additional-volumes"
						image = "ubuntu_focal"
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "additional_volumes" ]

						root_volume {
							size_in_gb = 15
						}
						additional_volume_ids = [
							scaleway_instance_volume.local.id,
							scaleway_block_volume.block.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					instancechecks.IsVolumePresent(tt, "scaleway_instance_volume.local"),
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.block"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-additional-volumes"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "15"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "additional_volume_ids.#", "2"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "additional_volume_ids.0", "scaleway_instance_volume.local", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "additional_volume_ids.1", "scaleway_block_volume.block", "id"),
				),
			},
			{
				// Detach volumes
				Config: `
					resource "scaleway_instance_volume" "local" {
						size_in_gb = 10
						type = "l_ssd"
					}

					resource "scaleway_block_volume" "block" {
						size_in_gb = 12
						iops = 5000
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-additional-volumes"
						image = "ubuntu_focal"
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "additional_volumes" ]

						root_volume {
							size_in_gb = 15
						}
						additional_volume_ids = []
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					instancechecks.IsVolumePresent(tt, "scaleway_instance_volume.local"),
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.block"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-additional-volumes"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "15"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "additional_volume_ids.#", "0"),
				),
			},
			{
				// With 2 additional blocks
				Config: `
					resource "scaleway_block_volume" "block" {
						size_in_gb = 12
						iops = 5000
					}

					resource "scaleway_block_volume" "bigger-volume" {
						iops = 15000
						size_in_gb = 75
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-additional-volumes"
						image = "ubuntu_focal"
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "additional_volumes" ]

						root_volume {
							size_in_gb = 15
						}
						additional_volume_ids = [
							scaleway_block_volume.block.id,
							scaleway_block_volume.bigger-volume.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.block"),
					blocktestfuncs.IsVolumePresent(tt, "scaleway_block_volume.bigger-volume"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-additional-volumes"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "15"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "additional_volume_ids.#", "2"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "additional_volume_ids.0", "scaleway_block_volume.block", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "additional_volume_ids.1", "scaleway_block_volume.bigger-volume", "id"),
				),
			},
		},
	})
}

func TestAccServer_ServerWithBlockNonDefaultZone(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_block_volume" "main" {
						iops       = 5000
						size_in_gb = 10
						zone       = "nl-ams-1"
					}

					resource "scaleway_block_volume" "additional" {
						iops       = 5000
						size_in_gb = 5
						zone       = "nl-ams-1"
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-with-block-non-default-zone"
						zone = "nl-ams-1"
						type = "DEV1-S"
						tags = [ "terraform-test", "scaleway_instance_server", "server_with_block_non_default_zone" ]

						root_volume {
							volume_id = scaleway_block_volume.main.id
							volume_type = "sbs_volume"
							delete_on_termination = true
						}
						additional_volume_ids = [scaleway_block_volume.additional.id]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "name", "tf-acc-server-with-block-non-default-zone"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "zone", "nl-ams-1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "root_volume.0.volume_id", "scaleway_block_volume.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.sbs_iops", "5000"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "additional_volume_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "additional_volume_ids.0", "scaleway_block_volume.additional", "id"),
				),
			},
		},
	})
}

func TestAccServer_ScratchStorage(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancechecks.IsServerDestroyed(tt),
			instancechecks.IsVolumeDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "main" {
						size_in_gb = 1600
						type = "scratch"
						zone = "fr-par-2"
					}
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-scratch-storage"
						type = "L40S-1-48G"
						image = "ubuntu_jammy_gpu_os_12"
						state = "stopped"
						zone = "fr-par-2"
					    tags  = [ "terraform-test", "scaleway_instance_server", "scratch_storage" ]
						additional_volume_ids = [scaleway_instance_volume.main.id]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.main"),
					instancechecks.IsVolumePresent(tt, "scaleway_instance_volume.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "name", "tf-acc-server-scratch-storage"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "L40S-1-48G"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "image", "ubuntu_jammy_gpu_os_12"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "additional_volume_ids.0", "scaleway_instance_volume.main", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.main", "size_in_gb", "1600"),
				),
			},
		},
	})
}
