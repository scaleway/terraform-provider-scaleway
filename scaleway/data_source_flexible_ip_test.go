package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceFlexibleIP_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_flexible_ip" "main" {
					}
					
					data "scaleway_flexible_ip" "by_address" {
						ip_address = "${scaleway_flexible_ip.main.ip_address}"
					}
					
					data "scaleway_flexible_ip" "by_id" {
						id = "${scaleway_flexible_ip.main.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.main"),

					resource.TestCheckResourceAttrSet("scaleway_flexible_ip.main", "ip_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_address", "ip_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_id", "ip_address"),

					resource.TestCheckResourceAttrSet("scaleway_flexible_ip.main", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_address", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_id", "id"),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceFlexibleIP_Multiple(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFlexibleIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_flexible_ip" "first" {
					}
					
					data "scaleway_flexible_ip" "by_address_first" {
						ip_address = "${scaleway_flexible_ip.first.ip_address}"
					}
					
					data "scaleway_flexible_ip" "by_id_first" {
						id = "${scaleway_flexible_ip.first.id}"
					}

					resource "scaleway_flexible_ip" "second" {
					}
					
					data "scaleway_flexible_ip" "by_address_second" {
						ip_address = "${scaleway_flexible_ip.second.ip_address}"
					}
					
					data "scaleway_flexible_ip" "by_id_second" {
						id = "${scaleway_flexible_ip.second.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.first"),

					resource.TestCheckResourceAttrSet("scaleway_flexible_ip.first", "ip_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_address_first", "ip_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_id_first", "ip_address"),

					resource.TestCheckResourceAttrSet("scaleway_flexible_ip.first", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_address_first", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_id_first", "id"),

					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.second"),

					resource.TestCheckResourceAttrSet("scaleway_flexible_ip.second", "ip_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_address_second", "ip_address"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_id_second", "ip_address"),

					resource.TestCheckResourceAttrSet("scaleway_flexible_ip.second", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_address_second", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_flexible_ip.by_id_second", "id"),
				),
			},
		},
	})
}
