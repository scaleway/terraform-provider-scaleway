package scaleway

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("scaleway_ip", &resource.Sweeper{
		Name: "scaleway_ip",
		F:    testSweepIP,
	})
}

func testSweepIP(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	scaleway := client.(*Client).scaleway
	log.Printf("[DEBUG] Destroying the ips in (%s)", region)

	ips, err := scaleway.GetIPS()
	if err != nil {
		return fmt.Errorf("Error describing IPs in Sweeper: %s", err)
	}

	for _, ip := range ips {
		if err := scaleway.DeleteIP(ip.ID); err != nil {
			return fmt.Errorf("Error deleting ip in Sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayIP_Count(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayIPDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewayIPConfig_Count,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPExists("scaleway_ip.base.0"),
					testAccCheckScalewayIPExists("scaleway_ip.base.1"),
				),
			},
		},
	})
}

func TestAccScalewayIP_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayIPDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckScalewayIPConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPExists("scaleway_ip.base"),
				),
			},
			resource.TestStep{
				Config: testAccCheckScalewayIPConfig_Reverse,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPExists("scaleway_ip.base"),
					resource.TestCheckResourceAttr(
						"scaleway_ip.base", "reverse", "www.google.de"),
				),
			},
			resource.TestStep{
				Config: testAccCheckScalewayIPAttachConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPExists("scaleway_ip.base"),
					testAccCheckScalewayIPAttachment("scaleway_ip.base", func(serverID string) bool {
						return serverID != ""
					}, "attachment failed"),
					resource.TestCheckResourceAttr(
						"scaleway_ip.base", "reverse", ""),
				),
			},
			resource.TestStep{
				Config: testAccCheckScalewayIPConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPExists("scaleway_ip.base"),
					testAccCheckScalewayIPAttachment("scaleway_ip.base", func(serverID string) bool {
						return serverID == ""
					}, "detachment failed"),
				),
			},
		},
	})
}

func testAccCheckScalewayIPDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).scaleway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway" {
			continue
		}

		_, err := client.GetIP(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("IP still exists")
		}
	}

	return nil
}

func testAccCheckScalewayIPAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return nil
	}
}

func testAccCheckScalewayIPExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IP ID is set")
		}

		client := testAccProvider.Meta().(*Client).scaleway
		ip, err := client.GetIP(rs.Primary.ID)

		if err != nil {
			return err
		}

		if ip.ID != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		return nil
	}
}

func testAccCheckScalewayIPAttachment(n string, check func(string) bool, msg string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IP ID is set")
		}

		client := testAccProvider.Meta().(*Client).scaleway
		ip, err := client.GetIP(rs.Primary.ID)

		if err != nil {
			return err
		}

		var serverID = ""
		if ip.Server != nil {
			serverID = ip.Server.Identifier
		}
		if !check(serverID) {
			return fmt.Errorf("IP check failed: %q", msg)
		}

		return nil
	}
}

var testAccCheckScalewayIPConfig_Reverse = `
resource "scaleway_ip" "base" {
  reverse = "www.google.de"
}
`

var testAccCheckScalewayIPConfig = `
resource "scaleway_ip" "base" {}
`

var testAccCheckScalewayIPConfig_Count = `
resource "scaleway_ip" "base" {
  count = 2
}
`

var testAccCheckScalewayIPAttachConfig = fmt.Sprintf(`
resource "scaleway_server" "base" {
  name = "test"
  # ubuntu 14.04
  image = "%s"
  type = "C1"
  state = "stopped"
}

resource "scaleway_ip" "base" {
  server = "${scaleway_server.base.id}"
}
`, armImageIdentifier)
