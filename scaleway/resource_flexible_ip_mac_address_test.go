package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
)

func TestAccScalewayFlexibleIPMACAddress_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "main" {}

						resource "scaleway_flexible_ip_mac_address" "main" {
						  flexible_ip_id = scaleway_flexible_ip.main.id
						  type = "kvm"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPAttachedMACAddress(tt, "scaleway_flexible_ip.main", "scaleway_flexible_ip_mac_address.main"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip_mac_address.main", "zone", "fr-par-1"),
					resource.TestCheckResourceAttr("scaleway_flexible_ip_mac_address.main", "type", "kvm"),
					resource.TestCheckResourceAttrSet("scaleway_flexible_ip_mac_address.main", "address"),
					resource.TestCheckResourceAttrSet("scaleway_flexible_ip_mac_address.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_flexible_ip_mac_address.main", "updated_at")),
			},
		},
	})
}

func TestAccScalewayFlexibleIPMACAddress_MoveToAnotherFlexibleIP(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_flexible_ip" "ip01" {}

						resource "scaleway_flexible_ip" "ip02" {}

						resource "scaleway_flexible_ip_mac_address" "main" {
						  flexible_ip_id = scaleway_flexible_ip.ip01.id
						  type = "kvm"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.ip01"),
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.ip02"),
					testAccCheckScalewayFlexibleIPAttachedMACAddress(tt, "scaleway_flexible_ip.ip01", "scaleway_flexible_ip_mac_address.main"),
					resource.TestCheckNoResourceAttr("scaleway_flexible_ip.ip02", "mac_address"),
				),
			},
			{
				Config: `
						resource "scaleway_flexible_ip" "ip01" {}

						resource "scaleway_flexible_ip" "ip02" {}

						resource "scaleway_flexible_ip_mac_address" "main" {
						  flexible_ip_id = scaleway_flexible_ip.ip02.id
						  type = "kvm"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.ip01"),
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.ip02"),
					testAccCheckScalewayFlexibleIPAttachedMACAddress(tt, "scaleway_flexible_ip.ip02", "scaleway_flexible_ip_mac_address.main"),
					resource.TestCheckNoResourceAttr("scaleway_flexible_ip.ip01", "mac_address"),
				),
			},
		},
	})
}

func TestAccScalewayFlexibleIPMACAddress_DuplicateOnOtherFlexibleIPs(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayFlexibleIPDestroy(tt),
			testAccCheckScalewayBaremetalServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
						data "scaleway_baremetal_offer" "my_offer" {
					      name = "EM-B112X-SSD"
				     	}

						resource "scaleway_baremetal_server" "base" {
						  name 			             = "TestAccScalewayBaremetalServer_WithoutInstallConfig"
						  offer     				 = data.scaleway_baremetal_offer.my_offer.offer_id
						  install_config_afterward   = true
						}
			
						resource "scaleway_flexible_ip" "ip01" {
                          server_id = scaleway_baremetal_server.base.id	
						}

						resource "scaleway_flexible_ip" "ip02" {
                          server_id = scaleway_baremetal_server.base.id
						}

						resource "scaleway_flexible_ip" "ip03" {
                          server_id = scaleway_baremetal_server.base.id
						}

						resource "scaleway_flexible_ip_mac_address" "main" {
						  flexible_ip_id = scaleway_flexible_ip.ip01.id
						  type           = "kvm"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.ip01"),
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.ip02"),
					testAccCheckScalewayFlexibleIPAttachedMACAddress(tt, "scaleway_flexible_ip.ip01", "scaleway_flexible_ip_mac_address.main"),
					resource.TestCheckNoResourceAttr("scaleway_flexible_ip.ip02", "mac_address"),
				),
			},
			{
				Config: `
						data "scaleway_baremetal_offer" "my_offer" {
					      name = "EM-B112X-SSD"
				     	}

						resource "scaleway_baremetal_server" "base" {
						  name 			             = "TestAccScalewayBaremetalServer_WithoutInstallConfig"
						  offer     				 = data.scaleway_baremetal_offer.my_offer.offer_id
						  install_config_afterward   = true
						}
			
						resource "scaleway_flexible_ip" "ip01" {
                          server_id = scaleway_baremetal_server.base.id	
						}

						resource "scaleway_flexible_ip" "ip02" {
                          server_id = scaleway_baremetal_server.base.id
						}

						resource "scaleway_flexible_ip" "ip03" {
                          server_id = scaleway_baremetal_server.base.id
						}

						resource "scaleway_flexible_ip_mac_address" "main" {
						  flexible_ip_id = scaleway_flexible_ip.ip01.id
						  type = "kvm"
                          flexible_ip_ids_to_duplicate = [
							scaleway_flexible_ip.ip02.id,
							scaleway_flexible_ip.ip03.id
						  ]
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.ip01"),
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.ip02"),
					testAccCheckScalewayFlexibleIPAttachedMACAddress(tt, "scaleway_flexible_ip.ip01", "scaleway_flexible_ip_mac_address.main"),
					testAccCheckScalewayFlexibleIPAttachedMACAddress(tt, "scaleway_flexible_ip.ip02", "scaleway_flexible_ip_mac_address.main"),
					testAccCheckScalewayFlexibleIPAttachedMACAddress(tt, "scaleway_flexible_ip.ip03", "scaleway_flexible_ip_mac_address.main"),
				),
			},
		},
	})
}

func testAccCheckScalewayFlexibleIPAttachedMACAddress(tt *TestTools, fipResource, macResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fipState, ok := s.RootModule().Resources[fipResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", fipResource)
		}
		macState, ok := s.RootModule().Resources[macResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", macResource)
		}

		fipAPI, zone, ID, err := fipAPIWithZoneAndID(tt.Meta, fipState.Primary.ID)
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

		if ip.MacAddress != nil && expandID(ip.MacAddress.ID) != expandID(macState.Primary.ID) {
			return fmt.Errorf("IDs should be the same in %s and %s: %v is different than %v", fipResource, macResource, ip.MacAddress.ID, macState.Primary.ID)
		}

		return nil
	}
}
