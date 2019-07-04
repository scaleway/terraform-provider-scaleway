package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func TestAccScalewayComputeInstanceIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayComputeInstanceIPConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.base"),
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.scaleway"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.base", "reverse", ""),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.scaleway", "reverse", "www.scaleway.com"),
				),
			},
			{
				Config: testAccScalewayComputeInstanceIPConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.base"),
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.scaleway"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.base", "reverse", "www.scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.scaleway", "reverse", ""),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceIP_Zone(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayComputeInstanceIPZoneConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: testAccScalewayComputeInstanceIPZoneConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.base", "zone", "nl-ams-1"),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceServerIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigIP("base1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base1"),
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base2"),
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.base_ip"),
					testAccCheckScalewayComputeInstanceIPPairWithServer("scaleway_compute_instance_ip.base_ip", "scaleway_compute_instance_server.base1"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigIP("base2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceIPPairWithServer("scaleway_compute_instance_ip.base_ip", "scaleway_compute_instance_server.base2"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigIP(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerNoIPAssigned("scaleway_compute_instance_server.base1"),
					testAccCheckScalewayComputeInstanceServerNoIPAssigned("scaleway_compute_instance_server.base2"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.base_ip", "server_id", ""),
				),
			},
		},
	})
}

func testAccCheckScalewayComputeInstanceIPExists(n string) resource.TestCheckFunc {
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
			IPID: ID,
			Zone: zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayComputeInstanceIPPairWithServer(ipResource, serverResource string) resource.TestCheckFunc {
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
			IPID: ID,
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

func testAccCheckScalewayComputeInstanceServerNoIPAssigned(serverResource string) resource.TestCheckFunc {
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

func testAccCheckScalewayComputeInstanceIPDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_compute_instance_ip" {
			continue
		}

		instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetIP(&instance.GetIPRequest{
			Zone: zone,
			IPID: ID,
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
var testAccScalewayComputeInstanceIPConfig = []string{
	`
		resource "scaleway_compute_instance_ip" "base" {}
		resource "scaleway_compute_instance_ip" "scaleway" {
			reverse = "www.scaleway.com"
		}
	`,
	`
		resource "scaleway_compute_instance_ip" "base" {
			reverse = "www.scaleway.com"	
		}
		resource "scaleway_compute_instance_ip" "scaleway" {}
	`,
}

// Check that we can change the zone of an ip (delete + create)
var testAccScalewayComputeInstanceIPZoneConfig = []string{
	`
		resource "scaleway_compute_instance_ip" "base" {}
	`,
	`
		resource "scaleway_compute_instance_ip" "base" {
			zone = "nl-ams-1"	
		}
	`,
}

func testAccCheckScalewayComputeInstanceServerConfigIP(attachedBase string) string {
	attachedServer := ""
	if attachedBase != "" {
		attachedServer = `server_id = "${scaleway_compute_instance_server.` + attachedBase + `.id}"`
	}
	return fmt.Sprintf(`
resource "scaleway_compute_instance_ip" "base_ip" {
  %s
}

resource "scaleway_compute_instance_server" "base1" {
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type  = "DEV1-S"
  
  tags  = [ "terraform-test", "scaleway_compute_instance_server", "attach_ip" ]
}

resource "scaleway_compute_instance_server" "base2" {
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type  = "DEV1-S"
  
  tags  = [ "terraform-test", "scaleway_compute_instance_server", "attach_ip" ]
}`, attachedServer)
}
