package instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

const (
	marketplaceImageType  = "instance_sbs"
	ubuntuFocalImageLabel = "ubuntu_focal"
	ubuntuJammyImageLabel = "ubuntu_jammy"
)

func TestAccServer_Minimal(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-minimal"
					  image = "ubuntu_focal"
					  type  = "DEV1-S"
					
					  tags = [ "terraform-test", "scaleway_instance_server", "minimal" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-minimal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "ubuntu_focal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "root_volume.0.volume_id"),
					serverHasNewVolume(tt, "scaleway_instance_server.base", "ubuntu_focal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "enable_dynamic_ip", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "minimal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "started"),
				),
			},
		},
	})
}

func TestAccServer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	serverID := new("")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-basic"
					  image = "ubuntu_focal"
					  type  = "DEV1-S"
					  tags = [ "terraform-test", "scaleway_instance_server", "basic" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					arePrivateNICsPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-basic"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "ubuntu_focal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.#", "3"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "basic"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", serverID),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-basic-renamed"
					  image = "ubuntu_focal"
					  type  = "DEV1-M"
					  tags = [ "terraform-test", "scaleway_instance_server", "basic", "more", "tags" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					arePrivateNICsPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-basic-renamed"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-M"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "ubuntu_focal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.#", "5"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "basic"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.3", "more"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.4", "tags"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", serverID),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-basic-renamed"
					  image = "ubuntu_focal"
					  type  = "DEV1-M"
					  tags = []
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					arePrivateNICsPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-basic-renamed"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "ubuntu_focal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-M"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.#", "0"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.base", serverID),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-basic-replaced"
					  image = "ubuntu_jammy"
					  type  = "DEV1-M"
					  tags = []
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					arePrivateNICsPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "tf-acc-server-basic-replaced"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "ubuntu_jammy"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-M"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.#", "0"),
					acctest.CheckResourceIDChanged("scaleway_instance_server.base", serverID), // changing image forces replacement
				),
			},
		},
	})
}

func TestAccServer_State_Stop(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// started
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-state-stop"
					  image = "ubuntu_jammy"
					  type  = "DEV1-S"
					  state = "started"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state_stop" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "started"),
				),
			},
			{
				// standby
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-state-stop"
					  image = "ubuntu_jammy"
					  type  = "DEV1-S"
					  state = "standby"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state_stop" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "standby"),
				),
			},
			{
				// stopped
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-state-stop"
					  image = "ubuntu_jammy"
					  type  = "DEV1-S"
					  state = "stopped"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state_stop" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "stopped"),
				),
			},
		},
	})
}

func TestAccServer_State_Start(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// stopped
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-state-start"
					  image = "ubuntu_jammy"
					  type  = "DEV1-S"
					  state = "stopped"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state_start" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "stopped"),
				),
			},
			{
				// standby
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-state-start"
					  image = "ubuntu_jammy"
					  type  = "DEV1-S"
					  state = "standby"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state_start" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "standby"),
				),
			},
			{
				// started
				Config: `
					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-state-start"
					  image = "ubuntu_jammy"
					  type  = "DEV1-S"
					  state = "started"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state_start" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "started"),
				),
			},
		},
	})
}

func TestAccServer_WithPlacementGroup(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_placement_group" "ha" {
						policy_mode = "enforced"
						policy_type = "max_availability"
					}

					resource "scaleway_instance_server" "ha" {
						count = 3
						name = "tf-acc-server-${count.index}-with-placement-group"
						image = "ubuntu_focal"
						type  = "PLAY2-PICO"
						placement_group_id = "${scaleway_instance_placement_group.ha.id}"
						tags  = [ "terraform-test", "scaleway_instance_server", "placement_group", "${count.index}" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.ha.0"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.ha.1"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.ha.2"),
					isPlacementGroupPresent(tt, "scaleway_instance_placement_group.ha"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.ha", "policy_respected", "true"),

					// placement_group_policy_respected is deprecated and should always be false.
					resource.TestCheckResourceAttr("scaleway_instance_server.ha.0", "placement_group_policy_respected", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_server.ha.1", "placement_group_policy_respected", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_server.ha.2", "placement_group_policy_respected", "false"),
				),
			},
		},
	})
}

func TestAccServer_CustomDiffImage(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	var mainServerID, controlServerID string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-custom-diff-image-main-server"
						image = "ubuntu_jammy"
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "custom_diff_image", "main" ]
					}
					resource "scaleway_instance_server" "control" {
						name = "tf-acc-server-custom-diff-image-control-server"
						image = "ubuntu_jammy"
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "custom_diff_image", "control" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.main"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.control"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "image", "ubuntu_jammy"),
					resource.TestCheckResourceAttr("scaleway_instance_server.control", "image", "ubuntu_jammy"),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.main", &mainServerID),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.control", &controlServerID),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_marketplace_image" "jammy" {
						label = "ubuntu_jammy"
						image_type = "%s"
					}
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-custom-diff-image-main-server"
						image = data.scaleway_marketplace_image.jammy.id
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "custom_diff_image", "main" ]
					}
					resource "scaleway_instance_server" "control" {
						name = "tf-acc-server-custom-diff-image-control-server"
						image = "ubuntu_jammy"
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "custom_diff_image", "control" ]
					}
				`, marketplaceImageType),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.main"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.control"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "image", "data.scaleway_marketplace_image.jammy", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.control", "image", "ubuntu_jammy"),
					imageIDMatchLabel(tt, "scaleway_instance_server.main", "scaleway_instance_server.control", true),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.main", &mainServerID),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.control", &controlServerID),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_marketplace_image" "focal" {
						label = "ubuntu_focal"
						image_type = "%s"
					}
					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-custom-diff-image-main-server"
						image = data.scaleway_marketplace_image.focal.id
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "custom_diff_image", "main" ]
					}
					resource "scaleway_instance_server" "control" {
						name = "tf-acc-server-custom-diff-image-control-server"
						image = "ubuntu_jammy"
						type = "DEV1-S"
						state = "stopped"
						tags = [ "terraform-test", "scaleway_instance_server", "custom_diff_image", "control" ]
					}
				`, marketplaceImageType),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.main"),
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.control"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "image", "data.scaleway_marketplace_image.focal", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.control", "image", "ubuntu_jammy"),
					imageIDMatchLabel(tt, "scaleway_instance_server.main", "scaleway_instance_server.control", false),
					acctest.CheckResourceIDChanged("scaleway_instance_server.main", &mainServerID),
					acctest.CheckResourceIDPersisted("scaleway_instance_server.control", &controlServerID),
				),
			},
		},
	})
}

func TestAccServer_AttachDetachFileSystem(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "terraform_instance_filesystem" {
						name = "filesystem-instance-terraform-test"
						size_in_gb = 100
						tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system", "fs01" ]
					}

					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-attach-detach-file-system"
					  type  = "POP2-HM-2C-16G"
					  state = "started"
					image = "%s"
					  tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system" ]
					  filesystems {
						filesystem_id = scaleway_file_filesystem.terraform_instance_filesystem.id
					  }
					}`, ubuntuJammyImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "POP2-HM-2C-16G"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "filesystems.0.filesystem_id"),
					serverHasNewVolume(tt, "scaleway_instance_server.base", ubuntuJammyImageLabel),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "attach_detach_file_system"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "terraform_instance_filesystem" {
						name = "filesystem-instance-terraform-test"
						size_in_gb = 100
						tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system", "fs01" ]
					}

					resource "scaleway_file_filesystem" "terraform_instance_filesystem_2" {
						name = "filesystem-instance-terraform-test-2"
						size_in_gb = 100
						tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system", "fs02" ]
					}

					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-attach-detach-file-system"
					  type  = "POP2-HM-2C-16G"
					  state = "started"
					  image = "%s"
					  tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system" ]

					   filesystems {
						filesystem_id = scaleway_file_filesystem.terraform_instance_filesystem_2.id
					  }
					}`, ubuntuJammyImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "POP2-HM-2C-16G"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "filesystems.0.filesystem_id"),
					resource.TestCheckNoResourceAttr("scaleway_instance_server.base", "filesystems.1.filesystem_id"),
					serverHasNewVolume(tt, "scaleway_instance_server.base", ubuntuJammyImageLabel),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "attach_detach_file_system"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_file_filesystem" "terraform_instance_filesystem" {
						name = "filesystem-instance-terraform-test"
						size_in_gb = 100
						tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system", "fs01" ]
					}

					resource "scaleway_file_filesystem" "terraform_instance_filesystem_2" {
						name = "filesystem-instance-terraform-test-2"
						size_in_gb = 100
						tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system", "fs02" ]
					}

					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-attach-detach-file-system"
					  type  = "POP2-HM-2C-16G"
					  state = "started"
 					  image = "%s"
					  tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system" ]
					  filesystems {
						filesystem_id = scaleway_file_filesystem.terraform_instance_filesystem_2.id
					  }
					  filesystems {
						filesystem_id = scaleway_file_filesystem.terraform_instance_filesystem.id
					  }
					}`, ubuntuJammyImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "POP2-HM-2C-16G"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "filesystems.0.filesystem_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "filesystems.1.filesystem_id"),
					serverHasNewVolume(tt, "scaleway_instance_server.base", ubuntuJammyImageLabel),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "attach_detach_file_system"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_block_volume" "volume" {
						iops = 15000
						size_in_gb = 15
					}

					resource "scaleway_file_filesystem" "terraform_instance_filesystem" {
						name = "filesystem-instance-terraform-test"
						size_in_gb = 100
						tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system", "fs01" ]
					}

					resource "scaleway_file_filesystem" "terraform_instance_filesystem_2" {
						name = "filesystem-instance-terraform-test-2"
						size_in_gb = 100
						tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system", "fs02" ]
					}

					resource "scaleway_instance_server" "base" {
					  name  = "tf-acc-server-attach-detach-file-system"
					  type  = "POP2-HM-2C-16G"
					  state = "started"
					  image = "%s"
					  tags  = [ "terraform-test", "scaleway_instance_server", "attach_detach_file_system" ]

					   filesystems {
						filesystem_id = scaleway_file_filesystem.terraform_instance_filesystem_2.id
					  }
					}`, ubuntuJammyImageLabel),
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "POP2-HM-2C-16G"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "filesystems.0.filesystem_id"),
					resource.TestCheckNoResourceAttr("scaleway_instance_server.base", "filesystems.1.filesystem_id"),
					serverHasNewVolume(tt, "scaleway_instance_server.base", ubuntuJammyImageLabel),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "attach_detach_file_system"),
				),
			},
		},
	})
}

func TestAccServer_AdminPasswordEncryptionSSHKeyID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	sshKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDFNaFderD6JUbMr6LoL7SdTaQ31gLcXwKv07Zyw0t4pq6Y8CGaeEvevS54TBR2iNJHa3hlIIUmA2qvH7Oh4v1QmMG2djWi2cD1lDEl8/8PYakaEBGh6snp3TMyhoqHOZqqKwDhPW0gJbe2vXfAgWSEzI8h1fs1D7iEkC1L/11hZjkqbUX/KduWFLyIRWdSuI3SWk4CXKRXwIkeYeSYb8AiIGY21u2z8H2J7YmhRzE85Kj/Fk4tST5gLW/IfLD4TMJjC/cZiJevETjs+XVmzTMIyU2sTQKufSQTj2qZ7RfgGwTHDoOeFvylgAdMGLZ/Un+gzeEPj9xUSPvvnbA9UPIKV4AffgtT1y5gcSWuHaqRxpUTY204mh6kq0EdVN2UsiJTgX+xnJgnOrKg6G3dkM8LSi2QtbjYbRXcuDJ9YUbUFK8M5Vo7LhMsMFb1hPtY68kbDUqD01RuMD5KhGIngCRRBZJriRQclUCJS4D3jr/Frw9ruNGh+NTIvIwdv0Y2brU= opensource@scaleway.com"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancechecks.IsServerDestroyed(tt),
			iamchecks.CheckSSHKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name = "test-acc-admin-pwd-encryption"
						public_key = %q
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-admin-password-encryption-ssh-key-id"
						type = "POP2-2C-8G-WIN"
						image = "windows_server_2022"
						tags  = [ "terraform-test", "scaleway_instance_server", "admin_password_encryption_ssh_key_id" ]
						admin_password_encryption_ssh_key_id = scaleway_iam_ssh_key.main.id
					}
					`, sshKey),
				Check: resource.ComposeTestCheckFunc(
					iamchecks.CheckSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "POP2-2C-8G-WIN"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "image", "windows_server_2022"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "admin_password_encryption_ssh_key_id", "scaleway_iam_ssh_key.main", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name = "test-acc-admin-pwd-encryption"
						public_key = %q
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-admin-password-encryption-ssh-key-id"
						type = "POP2-2C-8G-WIN"
						image = "windows_server_2022"
						tags  = [ "terraform-test", "scaleway_instance_server", "admin_password_encryption_ssh_key_id" ]
						admin_password_encryption_ssh_key_id = ""
					}
					`, sshKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "admin_password_encryption_ssh_key_id", ""),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name = "test-acc-admin-pwd-encryption"
						public_key = %q
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-admin-password-encryption-ssh-key-id"
						type = "POP2-2C-8G-WIN"
						image = "windows_server_2022"
						tags  = [ "terraform-test", "scaleway_instance_server", "admin_password_encryption_ssh_key_id" ]
						admin_password_encryption_ssh_key_id = scaleway_iam_ssh_key.main.id
					}
					`, sshKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "admin_password_encryption_ssh_key_id", "scaleway_iam_ssh_key.main", "id"),
				),
			},
		},
	})
}
