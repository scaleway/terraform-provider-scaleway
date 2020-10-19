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
		Name:         "scaleway_instance_ip",
		F:            testSweepInstanceIP,
		Dependencies: []string{"scaleway_instance_server"},
	})
}

func testSweepInstanceIP(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		instanceAPI := instance.NewAPI(scwClient)

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

func TestAccScalewayInstanceIP_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_instance_ip" "base" {}
						resource "scaleway_instance_ip" "scaleway" {}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists(tt, "scaleway_instance_ip.base"),
					testAccCheckScalewayInstanceIPExists(tt, "scaleway_instance_ip.scaleway"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceIP_WithZone(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_instance_ip" "base" {}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceIPExists(tt, "scaleway_instance_ip.base"),
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
					testAccCheckScalewayInstanceIPExists(tt, "scaleway_instance_ip.base"),
					resource.TestCheckResourceAttr("scaleway_instance_ip.base", "zone", "nl-ams-1"),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceIPExists(tt *TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
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

func testAccCheckScalewayInstanceIPPairWithServer(tt *TestTools, ipResource, serverResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ipState, ok := s.RootModule().Resources[ipResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", ipResource)
		}
		serverState, ok := s.RootModule().Resources[serverResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverResource)
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, ipState.Primary.ID)
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

func testAccCheckScalewayInstanceServerNoIPAssigned(tt *TestTools, serverResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		serverState, ok := s.RootModule().Resources[serverResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverResource)
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, serverState.Primary.ID)
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
			return fmt.Errorf("no flexible IP should be assigned to %s", serverResource)
		}

		return nil
	}
}

func testAccCheckScalewayInstanceIPDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_instance_ip" {
				continue
			}

			instanceAPI, zone, id, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = instanceAPI.GetIP(&instance.GetIPRequest{
				Zone: zone,
				IP:   id,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("resource %s(%s) still exist", rs.Type, rs.Primary.ID)
			}

			// Unexpected api error we return it
			// We check for 403 because instance API return 403 for deleted IP
			if !is404Error(err) && !is403Error(err) {
				return err
			}
		}

		return nil
	}
}
