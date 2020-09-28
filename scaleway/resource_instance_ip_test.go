package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_ip", &resource.Sweeper{
		Name: "scaleway_instance_ip",
		F:    testSweepInstanceIP,
	})
}

func testSweepInstanceIP(region string) error {
	return sweepZones(region, func(scwClient *scw.Client) error {
		instanceAPI := instance.NewAPI(scwClient)
		zone, _ := scwClient.GetDefaultZone()
		l.Debugf("sweeper: destroying the instance ip in (%s)", zone)
		listIPs, err := instanceAPI.ListIPs(&instance.ListIPsRequest{}, scw.WithAllPages())
		if err != nil {
			l.Warningf("error listing ips in (%s) in sweeper: %s", zone, err)
			return nil
		}

		for _, ip := range listIPs.IPs {
			err := instanceAPI.DeleteIP(&instance.DeleteIPRequest{
				IP: ip.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting ip in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayInstanceIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "base" {}
					resource "scaleway_instance_ip" "scaleway" {}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.base"),
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.scaleway"),
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
				Config: `
					resource "scaleway_instance_ip" "base" {}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.base"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "base" {
						zone = "nl-ams-1"	
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.base"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.base", "zone", "nl-ams-1"),
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

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
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

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(testAccProvider.Meta(), ipState.Primary.ID)
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

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(testAccProvider.Meta(), serverState.Primary.ID)
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

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
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
