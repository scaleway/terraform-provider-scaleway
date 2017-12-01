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
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	scaleway := client.(*Client).scaleway
	log.Printf("[DEBUG] Destroying the servers in (%s)", region)

	servers, err := scaleway.GetServers(true, 0)
	if err != nil {
		return fmt.Errorf("Error describing servers in Sweeper: %s", err)
	}

	for _, server := range *servers {
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewayServerConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerAttributes("scaleway_server.base"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "type", "C1"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "name", "test"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "tags.0", "terraform-test"),
				),
			},
			resource.TestStep{
				Config: testAccCheckScalewayServerConfig_IPAttachment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerIPAttachmentAttributes("scaleway_ip.base", "scaleway_server.base"),
				),
			},
			resource.TestStep{
				Config: testAccCheckScalewayServerConfig_IPDetachment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerIPDetachmentAttributes("scaleway_server.base"),
				),
			},
		},
	})
}

func TestAccScalewayServer_ExistingIP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewayServerVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerAttributes("scaleway_server.base"),
					resource.TestCheckResourceAttr(
						"scaleway_server.base", "type", "C1"),
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewayServerConfig_SecurityGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerSecurityGroup("scaleway_server.base", "blue"),
				),
			},
			resource.TestStep{
				Config: testAccCheckScalewayServerConfig_SecurityGroup_Update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerSecurityGroup("scaleway_server.base", "red"),
				),
			},
		},
	})
}

func TestAccScalewayServer_UserData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayServerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewayServerConfig_UserData,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerUserData("scaleway_server.base", map[string]string{
						"key1": "val1",
						"key2": "val2",
					}),
				),
			},
			resource.TestStep{
				Config: testAccCheckScalewayServerConfig_UserData_Update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayServerExists("scaleway_server.base"),
					testAccCheckScalewayServerUserData("scaleway_server.base", map[string]string{
						"key2": "val2",
						"key3": "val3",
					}),
				),
			},
		},
	})
}

func testAccCheckScalewayServerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).scaleway

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
		ip, ok := s.RootModule().Resources[ipName]
		if !ok {
			return fmt.Errorf("Unknown scaleway_ip resource: %s", ipName)
		}

		server, ok := s.RootModule().Resources[serverName]
		if !ok {
			return fmt.Errorf("Unknown scaleway_server resource: %s", serverName)
		}

		client := testAccProvider.Meta().(*Client).scaleway

		res, err := client.GetIP(ip.Primary.ID)
		if err != nil {
			return err
		}
		if res.IP.Server == nil || res.IP.Server.Identifier != server.Primary.ID {
			return fmt.Errorf("IP %q is not attached to server %q", ip.Primary.ID, server.Primary.ID)
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

		client := testAccProvider.Meta().(*Client).scaleway
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

func testAccCheckScalewayServerAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Unknown resource: %s", n)
		}

		client := testAccProvider.Meta().(*Client).scaleway
		server, err := client.GetServer(rs.Primary.ID)

		if err != nil {
			return err
		}

		if server.Name != "test" {
			return fmt.Errorf("Server has wrong name")
		}
		if server.Image.Identifier != armImageIdentifier {
			return fmt.Errorf("Wrong server image")
		}
		if server.CommercialType != "C1" {
			return fmt.Errorf("Wrong server type")
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

		client := testAccProvider.Meta().(*Client).scaleway
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

func testAccCheckScalewayServerUserData(n string, ud map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Unknown resource: %s", n)
		}

		client := testAccProvider.Meta().(*Client).scaleway
		aud, err := readUserDatas(client, rs.Primary.ID)
		if err != nil {
			return err
		}

		if len(ud) != len(aud) {
			return fmt.Errorf("Server has wrong number of user datas")
		}

		for k, v := range ud {
			ov, ok := aud[k]
			if !ok {
				return fmt.Errorf("Missing user data key: %s", k)
			}

			if v != ov {
				return fmt.Errorf("Wrong value for user data key %s: %s", k, ov)
			}
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

		client := testAccProvider.Meta().(*Client).scaleway
		server, err := client.GetServer(rs.Primary.ID)

		if err != nil {
			return err
		}

		if server.Identifier != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		return nil
	}
}

var armImageIdentifier = "5faef9cd-ea9b-4a63-9171-9e26bec03dbc"

var testAccCheckScalewayServerConfig = fmt.Sprintf(`
resource "scaleway_server" "base" {
  name = "test"
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  tags = [ "terraform-test" ]
}`, armImageIdentifier)

var testAccCheckScalewayServerConfig_IPAttachment = fmt.Sprintf(`
resource "scaleway_ip" "base" {}

resource "scaleway_server" "base" {
  name = "test"
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  tags = [ "terraform-test" ]
  public_ip = "${scaleway_ip.base.ip}"
}`, armImageIdentifier)

var testAccCheckScalewayServerConfig_IPDetachment = fmt.Sprintf(`
resource "scaleway_server" "base" {
  name = "test"
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  tags = [ "terraform-test" ]
}`, armImageIdentifier)

var testAccCheckScalewayServerVolumeConfig = fmt.Sprintf(`
resource "scaleway_server" "base" {
  name = "test"
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  tags = [ "terraform-test" ]

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
}`, armImageIdentifier)

var testAccCheckScalewayServerConfig_SecurityGroup = fmt.Sprintf(`
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
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  tags = [ "terraform-test" ]
  security_group = "${scaleway_security_group.blue.id}"
}`, armImageIdentifier)

var testAccCheckScalewayServerConfig_SecurityGroup_Update = fmt.Sprintf(`
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
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  tags = [ "terraform-test" ]
  security_group = "${scaleway_security_group.red.id}"
}`, armImageIdentifier)

var testAccCheckScalewayServerConfig_UserData = fmt.Sprintf(`
resource "scaleway_server" "base" {
  name = "test"
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  tags = [ "terraform-test" ]
	user_data {
		key1 = "val1"
		key2 = "val2"
	}
}`, armImageIdentifier)

var testAccCheckScalewayServerConfig_UserData_Update = fmt.Sprintf(`
resource "scaleway_server" "base" {
  name = "test"
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  tags = [ "terraform-test" ]
	user_data {
		key2 = "val2"
		key3 = "val3"
	}
}`, armImageIdentifier)
