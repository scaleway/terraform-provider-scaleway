package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayInstanceIPReverseDns_Basic(t *testing.T) {
	t.Skip("Skipping Reverse DNS because the domain is we can not execute dig +short www.scaleway-terraform.com ")
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "ip" {}
					resource "scaleway_instance_ip_reverse_dns" "base" {
						ip_id = scaleway_instance_ip.ip.id
						reverse = "www.scaleway-terraform.com"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_ip_reverse_dns.base", "reverse", "www.scaleway-terraform.com"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "ip" {}
				`,
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
