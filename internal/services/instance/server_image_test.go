package instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

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

func TestAccServer_ImageFromMarketplaceDataSource(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_marketplace_image" "local" {
						label         = "ubuntu_focal"
						instance_type = "DEV1-S"
					}

					resource "scaleway_instance_server" "main" {
						name  = "tf-acc-server-image-from-marketplace-datasource"
						image = "${data.scaleway_marketplace_image.local.id}"
						type  = "DEV1-S"
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.DoesImageExists(tt, "data.scaleway_marketplace_image.local"),
					resource.TestCheckResourceAttr("data.scaleway_marketplace_image.local", "label", "ubuntu_focal"),
					resource.TestCheckResourceAttr("data.scaleway_marketplace_image.local", "image_type", "instance_local"),
					resource.TestCheckResourceAttr("data.scaleway_marketplace_image.local", "instance_type", "DEV1-S"),
					resource.TestCheckResourceAttrPair("data.scaleway_marketplace_image.local", "id", "scaleway_instance_server.main", "image"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.volume_type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "image", "fr-par-1/"+ubuntuFocalLocalFrPar1ImageID),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "computed_image_id", ubuntuFocalLocalFrPar1ImageID),
				),
			},
			{
				Config: `
					data "scaleway_marketplace_image" "sbs" {
						label         = "ubuntu_focal"
						instance_type = "DEV1-S"
						image_type    = "instance_sbs"
					}

					resource "scaleway_instance_server" "main" {
						name  = "tf-acc-server-image-from-marketplace-datasource"
						image = "${data.scaleway_marketplace_image.sbs.id}"
						type  = "DEV1-S"
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.DoesImageExists(tt, "data.scaleway_marketplace_image.sbs"),
					resource.TestCheckResourceAttr("data.scaleway_marketplace_image.sbs", "label", "ubuntu_focal"),
					resource.TestCheckResourceAttr("data.scaleway_marketplace_image.sbs", "image_type", "instance_sbs"),
					resource.TestCheckResourceAttr("data.scaleway_marketplace_image.sbs", "instance_type", "DEV1-S"),
					resource.TestCheckResourceAttrPair("data.scaleway_marketplace_image.sbs", "id", "scaleway_instance_server.main", "image"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "root_volume.0.volume_type", "sbs_volume"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "image", "fr-par-1/"+ubuntuFocalSBSFrPar1ImageID),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "computed_image_id", ubuntuFocalSBSFrPar1ImageID),
				),
			},
		},
	})
}
