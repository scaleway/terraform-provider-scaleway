package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func TestAccScalewayInstanceIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayInstanceIPConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.base"),
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.scaleway"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.base", "reverse", ""),
					resource.TestCheckResourceAttr("scaleway_instance_ip.scaleway", "reverse", "www.scaleway.com"),
				),
			},
			{
				Config: testAccScalewayInstanceIPConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.base"),
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.scaleway"),
					// Do not work anymore because of scaleway_instance_ip_reverse_dns new resource.
					// Anyway the reverse attribute is deprecated.
					//resource.TestCheckResourceAttr("scaleway_instance_ip.base", "reverse", "www.scaleway.com"),
					//resource.TestCheckResourceAttr("scaleway_instance_ip.scaleway", "reverse", ""),
				),
			},
		},
	})
}

func TestAccScalewayInstanceIP_Zone(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayInstanceIPZoneConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.base"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: testAccScalewayInstanceIPZoneConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.base"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.base", "zone", "nl-ams-1"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigIP("base1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base1"),
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base2"),
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.base_ip"),
					testAccCheckScalewayInstanceIPPairWithServer("scaleway_instance_ip.base_ip", "scaleway_instance_server.base1"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigIP("base2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPPairWithServer("scaleway_instance_ip.base_ip", "scaleway_instance_server.base2"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigIP(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerNoIPAssigned("scaleway_instance_server.base1"),
					testAccCheckScalewayInstanceServerNoIPAssigned("scaleway_instance_server.base2"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.base_ip", "server_id", ""),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceIPExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetIP(&instance.GetIPRequest{
			IP:   ID,
			Zone: zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayInstanceIPPairWithServer(ipResource, serverResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ipState, ok := s.RootModule().Resources[ipResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", ipResource)
		}
		serverState, ok := s.RootModule().Resources[serverResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverResource)
		}

		instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), ipState.Primary.ID)
		if err != nil {
			return err
		}

		server, err := instanceAPI.GetServer(&instance.GetServerRequest{
			Zone:     zone,
			ServerID: expandID(serverState.Primary.ID),
		})
		if err != nil {
			return err
		}

		ip, err := instanceAPI.GetIP(&instance.GetIPRequest{
			IP:   ID,
			Zone: zone,
		})
		if err != nil {
			return err
		}

		if server.Server.PublicIP.Address.String() != ip.IP.Address.String() {
			return fmt.Errorf("IPs should be the same in %s and %s: %v is different than %v", ipResource, serverResource, server.Server.PublicIP.Address, ip.IP.Address)
		}

		return nil
	}
}

func testAccCheckScalewayInstanceServerNoIPAssigned(serverResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		serverState, ok := s.RootModule().Resources[serverResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverResource)
		}

		instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), serverState.Primary.ID)
		if err != nil {
			return err
		}

		server, err := instanceAPI.GetServer(&instance.GetServerRequest{
			Zone:     zone,
			ServerID: ID,
		})
		if err != nil {
			return err
		}

		if server.Server.PublicIP != nil && !server.Server.PublicIP.Dynamic {
			return fmt.Errorf("No flexible IP should be assigned to %s", serverResource)
		}

		return nil
	}
}

func testAccCheckScalewayInstanceIPDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_instance_ip" {
			continue
		}

		instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetIP(&instance.GetIPRequest{
			Zone: zone,
			IP:   ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("IP (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		// We check for 403 because instance API return 403 for deleted IP
		if !is404Error(err) && !is403Error(err) {
			return err
		}
	}

	return nil
}

// Check that reverse is handled at creation and update time
var testAccScalewayInstanceIPConfig = []string{
	`
		resource "scaleway_instance_ip" "base" {}
		resource "scaleway_instance_ip" "scaleway" {
			reverse = "www.scaleway.com"
		}
	`,
	`
		resource "scaleway_instance_ip" "base" {
			reverse = "www.scaleway.com"	
		}
		resource "scaleway_instance_ip" "scaleway" {}
	`,
}

// Check that we can change the zone of an ip (delete + create)
var testAccScalewayInstanceIPZoneConfig = []string{
	`
		resource "scaleway_instance_ip" "base" {}
	`,
	`
		resource "scaleway_instance_ip" "base" {
			zone = "nl-ams-1"	
		}
	`,
}

func testAccCheckScalewayInstanceServerConfigIP(attachedBase string) string {
	attachedServer := ""
	if attachedBase != "" {
		attachedServer = `server_id = "${scaleway_instance_server.` + attachedBase + `.id}"`
	}
	return fmt.Sprintf(`
resource "scaleway_instance_ip" "base_ip" {
  %s
}

resource "scaleway_instance_server" "base1" {
  image = "ubuntu-bionic"
  type  = "DEV1-S"
  
  tags  = [ "terraform-test", "scaleway_instance_server", "attach_ip" ]
}

resource "scaleway_instance_server" "base2" {
  image = "ubuntu-bionic"
  type  = "DEV1-S"
  
  tags  = [ "terraform-test", "scaleway_instance_server", "attach_ip" ]
}`, attachedServer)
}
