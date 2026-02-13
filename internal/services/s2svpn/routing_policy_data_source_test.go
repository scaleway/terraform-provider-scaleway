package s2svpn_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceRoutingPolicy_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckRoutingPolicyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_s2s_vpn_routing_policy" "main" {
						name             = "tf-test-routing-policy-ds"
						prefix_filter_in  = ["10.0.1.0/24"]
						prefix_filter_out = ["10.0.0.0/24"]
					}
				`,
			},
			{
				Config: `
					resource "scaleway_s2s_vpn_routing_policy" "main" {
						name             = "tf-test-routing-policy-ds"
						prefix_filter_in  = ["10.0.1.0/24"]
						prefix_filter_out = ["10.0.0.0/24"]
					}

					data "scaleway_s2s_vpn_routing_policy" "by_name" {
						name = scaleway_s2s_vpn_routing_policy.main.name
					}

					data "scaleway_s2s_vpn_routing_policy" "by_id" {
						routing_policy_id = scaleway_s2s_vpn_routing_policy.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingPolicyExists(tt, "scaleway_s2s_vpn_routing_policy.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_routing_policy.by_name", "name",
						"scaleway_s2s_vpn_routing_policy.main", "name"),
					resource.TestCheckResourceAttr(
						"data.scaleway_s2s_vpn_routing_policy.by_name",
						"prefix_filter_in.#", "1"),
					resource.TestCheckResourceAttr(
						"data.scaleway_s2s_vpn_routing_policy.by_name",
						"prefix_filter_out.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_routing_policy.by_id", "routing_policy_id",
						"scaleway_s2s_vpn_routing_policy.main", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_s2s_vpn_routing_policy.by_id", "name",
						"scaleway_s2s_vpn_routing_policy.main", "name"),
				),
			},
		},
	})
}
