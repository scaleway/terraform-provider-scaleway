package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceFlexibleIPs_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_flexible_ip" "first" {
						tags = [ "minimal", "first" ]
					}
				`,
			},
			{
				Config: `
					resource "scaleway_flexible_ip" "first" {
						tags = [ "minimal", "first" ]
					}
					
					resource "scaleway_flexible_ip" "second" {
						tags = [ "minimal", "second" ]
					}

					data "scaleway_flexible_ips" "fips_by_tags" {
						tags = [ "minimal" ]
						depends_on = [scaleway_flexible_ip.first, scaleway_flexible_ip.second]
					}

					data "scaleway_flexible_ips" "fips_by_tags_other_zone" {
						tags = [ "minimal" ]
						zone = "fr-par-2"
						depends_on = [scaleway_flexible_ip.first, scaleway_flexible_ip.second]
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ips.fips_by_tags", "ips.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ips.fips_by_tags", "ips.0.ip_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ips.fips_by_tags", "ips.1.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ips.fips_by_tags", "ips.1.ip_address"),

					resource.TestCheckNoResourceAttr("data.scaleway_flexible_ips.fips_by_tags_other_zone", "ips.0.id"),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceFlexibleIPs_WithBaremetalIDs(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbIPDestroy(tt),
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

					resource "scaleway_flexible_ip" "first" {
						tags = [ "minimal", "first" ]
					}
				`,
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

					resource "scaleway_flexible_ip" "first" {
						tags = [ "minimal", "first" ]
					}
					
					resource "scaleway_flexible_ip" "second" {
						tags = [ "minimal", "second" ]
					}
					`,
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

					resource "scaleway_flexible_ip" "first" {
						server_id = scaleway_baremetal_server.base.id
						tags = [ "minimal", "first" ]
					}
					
					resource "scaleway_flexible_ip" "second" {
						server_id = scaleway_baremetal_server.base.id
						tags = [ "minimal", "second" ]
					}

					data "scaleway_flexible_ips" "fips_by_tags" {
						server_ids = [scaleway_baremetal_server.base.id]
						depends_on = [scaleway_flexible_ip.first, scaleway_flexible_ip.second]
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ips.fips_by_tags", "ips.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ips.fips_by_tags", "ips.0.ip_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ips.fips_by_tags", "ips.1.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ips.fips_by_tags", "ips.1.ip_address"),
				),
			},
		},
	})
}
