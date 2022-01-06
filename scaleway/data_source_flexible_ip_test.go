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
						ip_id = "${scaleway_flexible_ip.main.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFlexibleIPExists(tt, "scaleway_flexible_ip.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_flexible_ip.by_address", "ip_address",
						"scaleway_flexible_ip.main", "ip_address",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_flexible_ip.by_id", "ip_id",
						"scaleway_flexible_ip.main", "ip_id",
					),
				),
			},
		},
	})
}
