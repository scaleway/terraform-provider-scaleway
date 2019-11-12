package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
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
				Check: resource.ComposeTestCheckFunc(),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "ip" {}
				`,
				Check: resource.ComposeTestCheckFunc(),
			},
		},
	})
}
