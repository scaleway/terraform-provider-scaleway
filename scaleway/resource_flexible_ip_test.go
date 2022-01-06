package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_flexible_ip", &resource.Sweeper{
		Name: "scaleway_flexible_ip",
		F:    testSweepFlexibleIP,
	})
}

func testSweepFlexibleIP(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		fipAPI := flexibleip.NewAPI(scwClient)

		listIPs, err := fipAPI.ListFlexibleIPs(&flexibleip.ListFlexibleIPsRequest{Zone: zone}, scw.WithAllPages())
		if err != nil {
			l.Warningf("error listing ips in (%s) in sweeper: %s", zone, err)
			return nil
		}

		for _, ip := range listIPs.FlexibleIPs {
			err := fipAPI.DeleteFlexibleIP(&flexibleip.DeleteFlexibleIPRequest{
				FipID: ip.ID,
				Zone:  zone,
			})
			if err != nil {
				return fmt.Errorf("error deleting ip in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayFlexibleIP_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "main" {}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
				),
			},
		},
	})
}

func TestAccScalewayFlexibleIP_WithZone(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {
							zone = "nl-ams-1"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "nl-ams-1"),
				),
			},
		},
	})
}

func TestAccScalewayFlexibleIP_CreateAndAttachToServer(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_instance_server" main {
							type = "DEV1-S"
							image = "ubuntu_focal"
							state = "stopped"
							enable_dynamic_ip = false
						}

						resource "scaleway_flexible_ip" "main" {
							zone = "fr-par-1"
							server_id = scaleway_instance_server.main.id
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					testAccCheckScalewayFlexibleIPAttachedToInstanceServer(tt, "scaleway_flexible_ip.main", "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.main", "zone", "fr-par-1"),
				),
			},
		},
	})
}

func TestAccScalewayFlexibleIP_AttachAndDetachFromServer(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_instance_server" main {
							type = "DEV1-S"
						}

						resource "scaleway_flexible_ip" "base" {
							zone = "fr-par-1"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: `
						resource "scaleway_instance_server" main {
							type = "DEV1-S"
						}

						resource "scaleway_flexible_ip" "main" {
							zone = "fr-par-1"
							server_id = scaleway_instance_server.main.id
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					testAccCheckScalewayFlexibleIPAttachedToInstanceServer(tt, "scaleway_flexible_ip.main", "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.main", "zone", "fr-par-1"),
				),
			},
			{
				Config: `
						resource "scaleway_instance_server" main {
							type = "DEV1-S"
						}

						resource "scaleway_flexible_ip" "main" {
							zone = "fr-par-1"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.main", "zone", "fr-par-1"),
				),
			},
		},
	})
}

func testAccCheckScalewayFlexibleIPExists(tt *TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		fipAPI, zone, ID, err := fipAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = fipAPI.GetFlexibleIP(&flexibleip.GetFlexibleIPRequest{
			FipID: ID,
			Zone:  zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayFlexibleIPDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_flexible_ip" {
				continue
			}

			fipAPI, zone, id, err := fipAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = fipAPI.GetFlexibleIP(&flexibleip.GetFlexibleIPRequest{
				FipID: id,
				Zone:  zone,
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

func testAccCheckScalewayFlexibleIPAttachedToInstanceServer(tt *TestTools, ipResource, serverResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ipState, ok := s.RootModule().Resources[ipResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", ipResource)
		}
		serverState, ok := s.RootModule().Resources[serverResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverResource)
		}

		instanceAPI, zone, _, err := instanceAPIWithZoneAndID(tt.Meta, ipState.Primary.ID)
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

		fipAPI, zone, ID, err := fipAPIWithZoneAndID(tt.Meta, ipState.Primary.ID)
		if err != nil {
			return err
		}
		ip, err := fipAPI.GetFlexibleIP(&flexibleip.GetFlexibleIPRequest{
			FipID: ID,
			Zone:  zone,
		})
		if err != nil {
			return err
		}

		if server.Server.PublicIP.Address.String() != ip.IPAddress.String() {
			return fmt.Errorf("IPs should be the same in %s and %s: %v is different than %v", ipResource, serverResource, server.Server.PublicIP.Address, ip.IPAddress)
		}

		return nil
	}
}
