package scaleway

import (
	"fmt"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
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

func TestAccScalewayFlexibleIP_CreateAndAttachToBaremetalServer(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalServer_Basic"
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIM7HUxRyQtB2rnlhQUcbDGCZcTJg7OvoznOiyC9W6IxH opensource@scaleway.com"
	name := "TestAccScalewayBaremetalServer_Basic"

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {
							zone = "fr-par-2"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-2"),
				),
			},
			{
				Config: fmt.Sprintf(`
						resource "scaleway_account_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
						}

						resource "scaleway_baremetal_server" "base" {
							name        = "%s"
							zone        = "fr-par-2"
							description = "test a description"
							offer       = "EM-A210R-HDD"
							os          = "96e5f0f2-d216-4de2-8a15-68730d877885"

							ssh_key_ids = [ scaleway_account_ssh_key.main.id ]
						}

						resource "scaleway_flexible_ip" "base" {
							server_id = scaleway_baremetal_server.base.id
							zone = "fr-par-2"
						}
					`, SSHKeyName, SSHKey, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					testAccCheckScalewayFlexibleIPAttachedToBaremetalServer(tt, "scaleway_flexible_ip.base", "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-2"),
				),
			},
		},
	})
}

// working with an already running baremetal called through data_source,
// but not working with the config below where a new baremetal is lanched from scratch
// having a POST error, status 500, giving up after 4 attempt(s)
// POST: https://api.scaleway.com/flexible-ip/v1alpha1/zones/fr-par-2/fips/detach
func TestAccScalewayFlexibleIP_AttachAndDetachFromBaremetalServer(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalServer_Basic"
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIM7HUxRyQtB2rnlhQUcbDGCZcTJg7OvoznOiyC9W6IxH opensource@scaleway.com"
	name := "TestAccScalewayBaremetalServer_Basic2"
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {
							zone = "fr-par-2"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-2"),
				),
			},
			{
				Config: fmt.Sprintf(`
						resource "scaleway_account_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
						}

						resource "scaleway_baremetal_server" "base" {
							name        = "%s"
							zone        = "fr-par-2"
							description = "test a description"
							offer       = "EM-A210R-HDD"
							os          = "96e5f0f2-d216-4de2-8a15-68730d877885"

							ssh_key_ids = [ scaleway_account_ssh_key.main.id ]
						}

						resource "scaleway_flexible_ip" "base" {
							server_id = scaleway_baremetal_server.base.id
							zone = "fr-par-2"
						}
					`, SSHKeyName, SSHKey, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					testAccCheckScalewayFlexibleIPAttachedToBaremetalServer(tt, "scaleway_flexible_ip.base", "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-2"),
				),
			},
			{
				Config: `
						resource "scaleway_flexible_ip" "base" {
							zone = "fr-par-2"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.base"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip.base", "zone", "fr-par-2"),
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

func testAccCheckScalewayFlexibleIPAttachedToBaremetalServer(tt *TestTools, ipResource, serverResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ipState, ok := s.RootModule().Resources[ipResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", ipResource)
		}
		serverState, ok := s.RootModule().Resources[serverResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", serverResource)
		}

		baremetalAPI, zoneID, err := baremetalAPIWithZoneAndID(tt.Meta, ipState.Primary.ID)
		if err != nil {
			return err
		}

		server, err := baremetalAPI.GetServer(&baremetal.GetServerRequest{
			Zone:     zoneID.Zone,
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

		if ip.ServerID == nil || server.ID != *ip.ServerID {
			return fmt.Errorf("IDs should be the same in %s and %s: %v is different than %v", ipResource, serverResource, server.ID, ip.ServerID)
		}

		return nil
	}
}
