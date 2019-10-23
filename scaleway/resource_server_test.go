package scaleway

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("scaleway_server", &resource.Sweeper{
		Name: "scaleway_server",
		F:    testSweepServer,
	})
}

func testSweepServer(region string) error {
	scaleway, err := sharedDeprecatedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	log.Printf("[DEBUG] Destroying the servers in (%s)", region)

	servers, err := scaleway.GetServers(true, 0)
	if err != nil {
		return fmt.Errorf("Error describing servers in Sweeper: %s", err)
	}

	for _, server := range servers {
		var err error
		if server.State == "stopped" {
			err = deleteStoppedServer(scaleway, &server)
		} else {
			err = deleteRunningServer(scaleway, &server)
		}

		if err != nil {
			return fmt.Errorf("Error deleting server in Sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayServer_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayServerConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "name", "test"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "boot_type", "local"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "cloudinit", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
			{
				Config: testAccCheckScalewayServerConfig_IPAttachment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerIPAttachmentAttributes("scaleway_ip.base", "scaleway_server.base"),
				),
			},
			{
				Config: testAccCheckScalewayServerConfig_IPDetachment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerIPDetachmentAttributes("scaleway_server.base"),
				),
			},
			{
				Config: testAccCheckScalewayServerConfig_dataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "state", "running"),
				),
			},
		},
	})
}

func TestAccScalewayServer_BootType(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayServerConfig_LocalBoot,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "boot_type", "local"),
				),
			},
		},
	})
}

func TestAccScalewayServer_ExistingIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayServerConfig_IPAttachment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerIPAttachmentAttributes("scaleway_ip.base", "scaleway_server.base"),
				),
			},
		},
	})
}

func TestAccScalewayServer_Volumes(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayServerVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "volume.#", "3"),
					resource.TestCheckResourceAttrSet(
						"scaleway_server.base", "volume.0.volume_id"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "volume.0.type", "l_ssd"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "volume.0.size_in_gb", "20"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "volume.1.type", "l_ssd"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "volume.1.size_in_gb", "0"),
					resource.TestCheckResourceAttrSet(
						"scaleway_server.base", "volume.2.volume_id"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "volume.2.type", "l_ssd"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "volume.2.size_in_gb", "30"),
				),
			},
		},
	})
}

func TestAccScalewayServer_SecurityGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayServerConfig_SecurityGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerSecurityGroup("scaleway_server.base", "blue"),
				),
			},
			{
				Config: testAccCheckScalewayServerConfig_SecurityGroup_Update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerSecurityGroup("scaleway_server.base", "red"),
				),
			},
		},
	})
}

func testAccCheckScalewayServerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Meta).deprecatedClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		_, err := client.GetServer(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Server still exists")
		}
	}

	return nil
}

func testAccCheckScalewayServerIPAttachmentAttributes(ipName, serverName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[ipName]
		if !ok {
			return fmt.Errorf("Unknown scaleway_ip resource: %s", ipName)
		}

		server, ok := s.RootModule().Resources[serverName]
		if !ok {
			return fmt.Errorf("Unknown scaleway_server resource: %s", serverName)
		}

		client := testAccProvider.Meta().(*Meta).deprecatedClient

		ip, err := client.GetIP(rs.Primary.ID)
		if err != nil {
			return err
		}
		if ip.Server == nil || ip.Server.Identifier != server.Primary.ID {
			return fmt.Errorf("IP %q is not attached to server %q", rs.Primary.ID, server.Primary.ID)
		}

		return nil
	}
}

func testAccCheckScalewayServerIPDetachmentAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Unknown resource: %s", n)
		}

		client := testAccProvider.Meta().(*Meta).deprecatedClient
		server, err := client.GetServer(rs.Primary.ID)
		if err != nil {
			return err
		}

		if server.PublicAddress.Identifier != "" {
			return fmt.Errorf("Expected server to have no public IP but got %q", server.PublicAddress.Identifier)
		}
		return nil
	}
}

func testAccCheckScalewayServerSecurityGroup(n, securityGroupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Unknown resource: %s", n)
		}

		client := testAccProvider.Meta().(*Meta).deprecatedClient
		server, err := client.GetServer(rs.Primary.ID)

		if err != nil {
			return err
		}

		if server.SecurityGroup.Name != securityGroupName {
			return fmt.Errorf("Server has wrong security_group")
		}

		return nil
	}
}

func testAccCheckScalewayServerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server ID is set")
		}

		client := testAccProvider.Meta().(*Meta).deprecatedClient
		server, err := client.GetServer(rs.Primary.ID)

		if err != nil {
			return err
		}

		if server.Identifier != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		if server.State != "running" {
			return fmt.Errorf("expected server to be running, but was %q", server.State)
		}

		return nil
	}
}

var testAccCheckScalewayServerConfig_dataSource = `
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  name         = "Ubuntu Bionic"
}

resource "scaleway_server" "base" {
  name = "test"

  image = "${data.scaleway_image.ubuntu.id}"
  type = "DEV1-S"
  tags = [ "terraform-test", "xenial" ]
}`

var testAccCheckScalewayServerConfig = `
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  name         = "Ubuntu Bionic"
  most_recent  = true
}

resource "scaleway_server" "base" {
  name = "test"
  image = "${data.scaleway_image.ubuntu.id}"
  type = "DEV1-S"
  tags = [ "terraform-test" ]
  cloudinit = <<EOF
#cloud-config
apt_update: true
apt_upgrade: true
EOF
}`

var testAccCheckScalewayServerConfig_LocalBoot = `
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  name         = "Ubuntu Xenial"
  most_recent  = true
}

resource "scaleway_server" "base" {
  name = "test"
  image = "${data.scaleway_image.ubuntu.id}"
  type = "START1-S" # This type is deprecated but as this resource will be removed in next major release we kept it.
  tags = [ "terraform-test", "local_boot" ]
  boot_type = "local"
}`

var testAccCheckScalewayServerConfig_IPAttachment = `
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  name         = "Ubuntu Bionic"
  most_recent  = true
}

resource "scaleway_ip" "base" {}

resource "scaleway_server" "base" {
  name = "test"
  image = "${data.scaleway_image.ubuntu.id}"
  type = "DEV1-S"
  tags = [ "terraform-test", "scaleway_ip" ]
  public_ip = "${scaleway_ip.base.ip}"
}`

var testAccCheckScalewayServerConfig_IPDetachment = `
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  most_recent  = true
  name         = "Ubuntu Bionic"
}

resource "scaleway_server" "base" {
  name = "test"
  image = "${data.scaleway_image.ubuntu.id}"
  type = "DEV1-S"
  tags = [ "terraform-test" ]
}
`

var testAccCheckScalewayServerVolumeConfig = `
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  name         = "Ubuntu Bionic"
  most_recent  = true
}

resource "scaleway_server" "base" {
  name = "test"
  image = "${data.scaleway_image.ubuntu.id}"
  type = "DEV1-S"
  tags = [ "terraform-test", "inline-images" ]

  volume {
    size_in_gb = 20
    type = "l_ssd"
  }

  volume {
    size_in_gb = 0
    type = "l_ssd"
  }

  volume {
    size_in_gb = 30
    type = "l_ssd"
  }
}`

var testAccCheckScalewayServerConfig_SecurityGroup = `
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  name         = "Ubuntu Bionic"
}

resource "scaleway_security_group" "blue" {
  name = "blue"
  description = "blue"
}

resource "scaleway_security_group" "red" {
  name = "red"
  description = "red"
}

resource "scaleway_server" "base" {
  name = "test"
  image = "${data.scaleway_image.ubuntu.id}"
  type = "DEV1-S"
  tags = [ "terraform-test", "security_groups.blue" ]
  security_group = "${scaleway_security_group.blue.id}"
}`

var testAccCheckScalewayServerConfig_SecurityGroup_Update = `
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  name         = "Ubuntu Bionic"
  most_recent  = true
}

resource "scaleway_security_group" "blue" {
  name = "blue"
  description = "blue"
}

resource "scaleway_security_group" "red" {
  name = "red"
  description = "red"
}

resource "scaleway_server" "base" {
  name = "test"
  image = "${data.scaleway_image.ubuntu.id}"
  type = "DEV1-S"
  tags = [ "terraform-test", "security_groups.red" ]
  security_group = "${scaleway_security_group.red.id}"
}`
