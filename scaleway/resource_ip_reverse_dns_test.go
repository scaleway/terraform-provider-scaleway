package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalewayIPReverseDNS_Basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayIPReverseDNSConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPExists("scaleway_ip.base"),
					resource.TestCheckResourceAttr(
						"scaleway_ip_reverse_dns.google", "reverse", "www.google.com"),
				),
			},
			{
				Config: testAccCheckScalewayIPReverseDNSConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIPExists("scaleway_ip.base"),
					resource.TestCheckResourceAttr(
						"scaleway_ip.base", "reverse", "www.google.com"),
				),
			},
		},
	})
}

var testAccCheckScalewayIPReverseDNSConfig = `
resource "scaleway_ip" "base" {}

resource "scaleway_ip_reverse_dns" "google" {
	ip = "${scaleway_ip.base.id}"
	reverse = "www.google.com"
}
`
