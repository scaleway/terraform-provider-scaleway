package s2svpn_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	s2s_vpn "github.com/scaleway/scaleway-sdk-go/api/s2s_vpn/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/s2svpn"
)

func TestAccRoutingPolicy_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckRoutingPolicyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_s2s_vpn_routing_policy" "main" {
						name              = "tf-test-routing-policy"
						prefix_filter_in  = ["10.0.1.0/24"]
						prefix_filter_out = ["10.0.2.0/24"]
						region            = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingPolicyExists(tt, "scaleway_s2s_vpn_routing_policy.main"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_routing_policy.main", "name", "tf-test-routing-policy"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_routing_policy.main", "is_ipv6", "false"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_routing_policy.main", "region", "fr-par"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_routing_policy.main", "prefix_filter_in.#", "1"),
					resource.TestCheckResourceAttr("scaleway_s2s_vpn_routing_policy.main", "prefix_filter_out.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_routing_policy.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_s2s_vpn_routing_policy.main", "updated_at"),
				),
			},
		},
	})
}

func testAccCheckRoutingPolicyExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := s2svpn.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetRoutingPolicy(&s2s_vpn.GetRoutingPolicyRequest{
			RoutingPolicyID: id,
			Region:          region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckRoutingPolicyDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_s2s_vpn_routing_policy" {
				continue
			}

			api, region, id, err := s2svpn.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetRoutingPolicy(&s2s_vpn.GetRoutingPolicyRequest{
				RoutingPolicyID: id,
				Region:          region,
			})
			if err == nil {
				return fmt.Errorf("routing policy (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
