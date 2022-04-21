package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayInstanceIPReverseDns_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	testDNSZone := fmt.Sprintf("tf.%s", testDomain)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_ip" "main" {}
					
					resource "scaleway_domain_record" "tf_A" {
						dns_zone = %[1]q
						name     = "tf"
						type     = "A"
						data     = "${scaleway_instance_ip.main.address}"
						ttl      = 3600
						priority = 1
					}

					resource "scaleway_instance_ip_reverse_dns" "base" {
						ip_id = scaleway_instance_ip.main.id
						reverse = %[2]q
					}
				`, testDomain, testDNSZone),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_ip_reverse_dns.base", "reverse", testDNSZone),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "ip" {}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_ip.ip", "reverse", ""),
				),
			},
		},
	})
}
