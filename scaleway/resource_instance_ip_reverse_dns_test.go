package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayInstanceReverseDnsIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "ip" {}
					resource "scaleway_instance_ip_reverse_dns" "base" {
						ip_id = scaleway_instance_ip.ip.id
						reverse = "www.google.fr"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_ip_reverse_dns.base", "reverse", "www.google.fr"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "ip" {}
				`,
				Taint: []string{"scaleway_instance_ip.ip"},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_ip.ip", "reverse", ""),
				),
			},
		},
	})
}
