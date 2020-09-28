package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccScalewayVolumeAttachment_Basic(t *testing.T) {
	t.Skip("C2S instance are EOL")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayVolumeAttachmentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayVolumeAttachmentExists("scaleway_volume_attachment.test"),
				),
			},
		},
	})
}

func testAccCheckScalewayVolumeAttachmentDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Meta).deprecatedClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		s, err := client.GetServer(rs.Primary.Attributes["server"])
		if err != nil {
			fmt.Printf("Failed getting server: %q", err)
			return err
		}

		for _, volume := range s.Volumes {
			if volume.Identifier == rs.Primary.Attributes["volume"] {
				return fmt.Errorf("Attachment still exists")
			}
		}
	}

	return nil
}

func testAccCheckScalewayVolumeAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Meta).deprecatedClient

		rs := s.RootModule().Resources[n]

		server, err := client.GetServer(rs.Primary.Attributes["server"])
		if err != nil {
			fmt.Printf("Failed getting server: %q", err)
			return err
		}

		for _, volume := range server.Volumes {
			if volume.Identifier == rs.Primary.Attributes["volume"] {
				return nil
			}
		}

		return fmt.Errorf("Attachment does not exist")
	}
}

var testAccCheckScalewayVolumeAttachmentConfig = `
data "scaleway_image" "ubuntu" {
  name         = "Ubuntu Xenial"
  architecture = "x86_64"
}

resource "scaleway_server" "base" {
  name = "test"
  image = "${data.scaleway_image.ubuntu.id}"
  type = "C2S"

  tags = [ "terraform-test", "external-volume-attachment" ]
}

resource "scaleway_volume" "test" {
  name = "test"
  size_in_gb = 5
  type = "l_ssd"
}

resource "scaleway_volume_attachment" "test" {
  server = "${scaleway_server.base.id}"
  volume = "${scaleway_volume.test.id}"
}`
