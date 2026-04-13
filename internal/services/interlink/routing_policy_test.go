package interlink_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	interlinkSDK "github.com/scaleway/scaleway-sdk-go/api/interlink/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/interlink"
)

func TestAccInterlinkRoutingPolicy_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckInterlinkRoutingPolicyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_interlink_routing_policy" "main" {
						name              = "tf-test-interlink-routing-policy"
						prefix_filter_in  = ["10.0.1.0/24"]
						prefix_filter_out = ["10.0.2.0/24"]
						region            = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInterlinkRoutingPolicyExists(tt, "scaleway_interlink_routing_policy.main"),
					resource.TestCheckResourceAttr("scaleway_interlink_routing_policy.main", "name", "tf-test-interlink-routing-policy"),
					resource.TestCheckResourceAttr("scaleway_interlink_routing_policy.main", "is_ipv6", "false"),
					resource.TestCheckResourceAttr("scaleway_interlink_routing_policy.main", "region", "fr-par"),
					resource.TestCheckResourceAttr("scaleway_interlink_routing_policy.main", "prefix_filter_in.#", "1"),
					resource.TestCheckResourceAttr("scaleway_interlink_routing_policy.main", "prefix_filter_out.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_interlink_routing_policy.main", "created_at"),
					resource.TestCheckResourceAttrSet("scaleway_interlink_routing_policy.main", "updated_at"),
				),
			},
			{
				ResourceName:      "scaleway_interlink_routing_policy.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckInterlinkRoutingPolicyExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := interlink.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetRoutingPolicy(&interlinkSDK.GetRoutingPolicyRequest{
			RoutingPolicyID: id,
			Region:          region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckInterlinkRoutingPolicyDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_interlink_routing_policy" {
				continue
			}

			api, region, id, err := interlink.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetRoutingPolicy(&interlinkSDK.GetRoutingPolicyRequest{
				RoutingPolicyID: id,
				Region:          region,
			})
			if err == nil {
				return fmt.Errorf("interlink routing policy (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
