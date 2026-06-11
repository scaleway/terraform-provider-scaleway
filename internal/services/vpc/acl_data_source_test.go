package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceACL_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isACLDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc01" {
					  name = "tf-vpc-ds-acl"
					}

					resource "scaleway_vpc_acl" "acl01" {
					  vpc_id         = scaleway_vpc.vpc01.id
					  is_ipv6        = false
					  default_policy = "drop"
					  rules {
					    protocol      = "TCP"
					    src_port_low  = 0
					    src_port_high = 0
					    dst_port_low  = 80
					    dst_port_high = 80
					    source        = "0.0.0.0/0"
					    destination   = "0.0.0.0/0"
					    description   = "Allow HTTP"
					    action        = "accept"
					  }
					}

					data "scaleway_vpc_acl" "by_vpc_id" {
					  vpc_id     = scaleway_vpc.vpc01.id
					  is_ipv6    = false
					  depends_on = [scaleway_vpc_acl.acl01]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isACLPresent(tt, "scaleway_vpc_acl.acl01"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_acl.by_vpc_id", "default_policy", "drop"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_acl.by_vpc_id", "rules.#", "1"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_acl.by_vpc_id", "rules.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_acl.by_vpc_id", "rules.0.dst_port_low", "80"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_acl.by_vpc_id", "rules.0.dst_port_high", "80"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_acl.by_vpc_id", "rules.0.action", "accept"),
					resource.TestCheckResourceAttr("data.scaleway_vpc_acl.by_vpc_id", "rules.0.description", "Allow HTTP"),
				),
			},
		},
	})
}
